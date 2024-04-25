package tax

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type Handler struct {
	store Storer
}

type Storer interface {
	TaxRatesIncome(finalIncome float64) (TaxRate, error)
	TaxRates() ([]TaxRate, error)
	TaxDeductionByType(allowanceTypes []string) ([]TaxDeduction, error)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func (h *Handler) CalculationHandler(c echo.Context) error {

	payload := new(TaxCalculation)

	err := c.Bind(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	allowanceType := []string{}

	for _, v := range payload.Allowances {
		allowanceType = append(allowanceType, v.AllowanceType)
	}

	allowanceType = append(allowanceType, "personal")
	taxDeductions, err := h.store.TaxDeductionByType(allowanceType)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	err = validation(taxDeductions, *payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	deducted := payload.TotalIncome

	maxDeduct := maxDeduct(taxDeductions, payload.Allowances)
	deducted -= maxDeduct

	taxRates, err := h.store.TaxRates()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	foundKey := slices.IndexFunc(taxRates, func(t TaxRate) bool {
		return deducted <= t.LowerBoundIncome
	})

	rIndex := foundKey

	if foundKey > 0 {
		rIndex = foundKey - 1
	} else {
		rIndex = 0
	}

	taxFund := calculateTaxPayable(deducted, payload.WithHoldingTax, taxRates[rIndex])

	taxRefund := 0.0
	var taxLevels []TaxLevelInfo
	p := message.NewPrinter(language.English)

	if math.Signbit(taxFund) {
		taxRefund = math.Abs(taxFund)
		taxFund = 0.0
	}

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

	return c.JSON(http.StatusOK, CalculationResponse{
		Tax:       taxFund,
		TaxRefund: taxRefund,
		TaxLevel:  taxLevels,
	})
}

func (h *Handler) CalculationCSV(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	fileUploaded, err := file.Open()
	if err != nil {
		return err
	}

	defer fileUploaded.Close()

	calculateFromFile(fileUploaded)
	return c.JSON(http.StatusOK, "OK")
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

func calculateTaxPayable(deducted float64, wht float64, rate TaxRate) float64 {

	baseTax := 150000.0
	taxPercent := (rate.TaxRate / 100)
	taxfund := ((deducted - baseTax) * taxPercent) - wht
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

func calculateFromFile(file io.Reader) error {
	read := csv.NewReader(file)

	recs, err := read.ReadAll()
	if err != nil {
		return err
	}

	for i, v := range recs {
		fmt.Println("record %d : %v", i, v)
	}
	return nil
}
