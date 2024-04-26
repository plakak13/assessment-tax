package tax

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/plakak13/assessment-tax/helper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type Handler struct {
	store Storer
}

type Storer interface {
	TaxRates() ([]TaxRate, error)
	TaxDeductionByType(allowanceTypes []string) ([]TaxDeduction, error)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func (h *Handler) CalculationHandler(c echo.Context) error {

	tc := new(TaxCalculation)

	err := c.Bind(tc)
	if err != nil {
		return helper.FailedHandler(c, err.Error(), http.StatusBadRequest)
	}
	allowanceType := []string{}

	for _, v := range tc.Allowances {
		allowanceType = append(allowanceType, v.AllowanceType)
	}

	allowanceType = append(allowanceType, "personal")
	tds, err := h.store.TaxDeductionByType(allowanceType)

	if err != nil {
		return helper.FailedHandler(c, err.Error())
	}

	if err = validation(tds, *tc); err != nil {
		return helper.FailedHandler(c, err.Error(), http.StatusBadRequest)
	}

	income := tc.TotalIncome

	maxDeduct := maxDeduct(tds, tc.Allowances)
	income -= maxDeduct

	taxRates, err := h.store.TaxRates()

	if err != nil {
		return helper.FailedHandler(c, err.Error())
	}

	rIndex := taxRateIndex(taxRates, income)

	taxFund := calculateTaxPayable(income, tc.WithHoldingTax, taxRates[rIndex])

	taxRefund, taxFund := refundTax(taxFund)
	taxLevels := taxLevelDetails(taxRates, rIndex, taxFund)

	return helper.SuccessHandler(c, CalculationResponse{
		Tax:       math.Round(taxFund*100) / 100,
		TaxRefund: math.Round(taxRefund*100) / 100,
		TaxLevel:  taxLevels,
	})
}

func (h *Handler) CalculationCSV(c echo.Context) error {

	fileUploaded, err := openFile(c)
	if err != nil {
		return helper.FailedHandler(c, err.Error(), http.StatusBadRequest)
	}

	recs, err := readCSVRecords(fileUploaded)
	if err != nil {
		return helper.FailedHandler(c, err.Error(), http.StatusBadRequest)
	}

	recs = removeBOM(recs)

	if !validateCSVHeader(recs[0]) {
		return helper.FailedHandler(c, "header incorrect", http.StatusBadRequest)
	}

	allowanceType := []string{"personal", "donation"}

	tds, err := h.store.TaxDeductionByType(allowanceType)
	if err != nil {
		return helper.FailedHandler(c, err.Error())
	}
	var ttis []TaxWithTotalIncome

	for i, v := range recs {
		if i == 0 {
			continue
		}

		totalIncome, wht, amount, err := parseFloatValue(v)
		if err != nil {
			return helper.FailedHandler(c, err.Error(), http.StatusBadRequest)
		}
		tc := TaxCalculation{
			TotalIncome:    totalIncome,
			WithHoldingTax: wht,
			Allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        amount,
				},
			},
		}

		if err = validation(tds, tc); err != nil {
			return helper.FailedHandler(c, err.Error(), http.StatusBadRequest)
		}

		maxDeduct := maxDeduct(tds, []Allowance{
			{
				AllowanceType: "donation",
				Amount:        amount,
			},
		})

		income := totalIncome - maxDeduct

		taxRates, err := h.store.TaxRates()

		if err != nil {
			return helper.FailedHandler(c, err.Error())
		}

		rIndex := taxRateIndex(taxRates, income)

		taxFund := calculateTaxPayable(income, wht, taxRates[rIndex])

		ttis = append(ttis, TaxWithTotalIncome{
			TotalIncome: totalIncome,
			TaxAmount:   taxFund,
		})
	}

	return helper.SuccessHandler(c, ttis)
}

func validation(taxDeducts []TaxDeduction, t TaxCalculation) error {

	if t.WithHoldingTax <= 0 || t.WithHoldingTax > t.TotalIncome {
		return errors.New("invalid withholding tax amount")
	}

	for _, v := range t.Allowances {
		for _, vt := range taxDeducts {
			if vt.TaxAllowanceType == v.AllowanceType && v.Amount < vt.MinAmount {
				return fmt.Errorf("amount for %s allowance is below the minimum threshold", v.AllowanceType)
			}
		}
	}
	return nil
}

func calculateTaxPayable(income float64, wht float64, rate TaxRate) float64 {

	baseTax := 150000.0
	taxPercent := (rate.TaxRate / 100)
	taxfund := ((income - baseTax) * taxPercent) - wht
	return taxfund
}

func maxDeduct(tds []TaxDeduction, alls []Allowance) float64 {
	var maxDeduct float64
	for _, td := range tds {
		if td.TaxAllowanceType == "personal" {
			maxDeduct += td.MaxDeductionAmount
		}
		for _, a := range alls {
			if td.TaxAllowanceType == a.AllowanceType {
				if td.MaxDeductionAmount >= a.Amount {
					maxDeduct += a.Amount
				} else {
					maxDeduct += td.MaxDeductionAmount
				}
			}
		}
	}
	return maxDeduct
}

func validateCSVHeader(h []string) bool {
	expected := []string{"totalIncome", "wht", "donation"}
	return equalSlice(expected, h)
}

func equalSlice(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func praseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func openFile(c echo.Context) (multipart.File, error) {

	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}

	fileUploaded, err := file.Open()
	if err != nil {
		return nil, err
	}

	defer fileUploaded.Close()
	return fileUploaded, nil
}

func removeBOM(recs [][]string) [][]string {
	if len(recs) > 0 && strings.HasPrefix(recs[0][0], "\ufeff") {
		recs[0][0] = strings.TrimPrefix(recs[0][0], "\ufeff")
	}
	return recs
}

func taxLevelDetails(taxRates []TaxRate, rIndex int, taxFund float64) []TaxLevelInfo {
	var taxLevels []TaxLevelInfo
	p := message.NewPrinter(language.English)
	for i, v := range taxRates {
		tVal := 0.0
		if v.ID == taxRates[rIndex].ID {
			tVal = taxFund
		}
		if i+1 != len(taxRates) {
			tFormat := p.Sprintf("%v-%v", number.Decimal(v.LowerBoundIncome), number.Decimal(taxRates[i+1].LowerBoundIncome-1))

			taxLevels = append(taxLevels, TaxLevelInfo{
				Tax:   tVal,
				Level: tFormat,
			})

		} else {
			lastT := p.Sprintf("%v ขึ้นไป", number.Decimal(v.LowerBoundIncome))

			taxLevels = append(taxLevels, TaxLevelInfo{
				Tax:   tVal,
				Level: lastT,
			})

		}
	}
	return taxLevels
}

func refundTax(taxFund float64) (float64, float64) {
	taxRefund := 0.0

	if math.Signbit(taxFund) {
		taxRefund = math.Abs(taxFund)
		taxFund = 0.0
	}
	return taxRefund, taxFund
}

func taxRateIndex(taxRates []TaxRate, income float64) int {

	foundKey := slices.IndexFunc(taxRates, func(t TaxRate) bool {
		return income <= t.LowerBoundIncome
	})

	rIndex := foundKey

	if foundKey > 0 {
		rIndex = foundKey - 1
	} else {
		rIndex = 0
	}
	return rIndex
}

func readCSVRecords(fileUploaded io.Reader) ([][]string, error) {
	read := csv.NewReader(fileUploaded)

	recs, err := read.ReadAll()
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func parseFloatValue(v []string) (float64, float64, float64, error) {
	totalIncome, err := praseFloat(v[0])
	if err != nil {
		return 0, 0, 0, errors.New("total income can not be string or empty")
	}

	wht, err := praseFloat(v[1])
	if err != nil {
		return 0, 0, 0, errors.New("tax with holding (twh) can not be string or empty")
	}

	amount, err := praseFloat(v[2])
	if err != nil {
		return 0, 0, 0, errors.New("donation can not be string or empty")
	}

	return totalIncome, wht, amount, nil
}
