package tax

import (
	"errors"
	"fmt"
	"net/http"

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

	taxRate, err := h.store.TaxRatesIncome(deducted)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	taxFund := calculateTaxPayable(deducted, payload.WithHoldingTax, taxRate)

	taxRates, err := h.store.TaxRates()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	var taxLevels []TaxLevelInfo
	p := message.NewPrinter(language.English)

	for i, v := range taxRates {
		fTaz := 0.0
		if v.ID == taxRate.ID {
			fTaz = taxFund
		}
		if i+1 != len(taxRates) {
			tFormat := p.Sprintf("%v-%v", number.Decimal(v.LowerBoundIncome), number.Decimal(taxRates[i+1].LowerBoundIncome-1))

			taxLevels = append(taxLevels, TaxLevelInfo{
				Tax:   fTaz,
				Level: tFormat,
			})

		} else {
			lastT := p.Sprintf("%v ขึ้นไป", number.Decimal(v.LowerBoundIncome))

			taxLevels = append(taxLevels, TaxLevelInfo{
				Tax:   fTaz,
				Level: lastT,
			})

		}

	}

	return c.JSON(http.StatusOK, CalculationResponse{Tax: taxFund, TaxLevel: taxLevels})
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
