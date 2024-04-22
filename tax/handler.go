package tax

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface {
	TaxRates(finalIncome float64) (TaxRate, error)
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
		fmt.Println("error===>" + err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pass := validation(taxDeductions, *payload)
	if !pass {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	deduct := payload.TotalIncome

	maxDeduct := maxDeduct(taxDeductions, payload.Allowances)
	deduct -= maxDeduct
	taxRate, err := h.store.TaxRates(deduct)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	taxFund := calculateTaxPayable(deduct, payload.WithHoldingTax, taxRate)

	return c.JSON(http.StatusOK, CalculationResponse{Tax: taxFund})
}

func validation(taxDeducts []TaxDeduction, t TaxCalculation) bool {

	if t.WithHoldingTax <= 0 || t.WithHoldingTax > t.TotalIncome {
		return false
	}

	for _, v := range t.Allowances {
		for _, vt := range taxDeducts {
			if vt.TaxAllowanceType == v.AllowanceType && v.Amount < vt.MinAmount {
				return false
			}
		}
	}
	return true
}

func calculateTaxPayable(deduct float64, wht float64, rate TaxRate) float64 {

	baseTax := 150000.0
	taxPercent := (rate.TaxRate / 100)
	taxfund := ((deduct - baseTax) * taxPercent) - wht
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
