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

	deduct := payload.TotalIncome - payload.WithHoldingTax
	for _, v := range payload.Allowances {
		deduct -= v.Amount
	}

	taxRate, err := h.store.TaxRates(deduct)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	taxFund := calculate(deduct, taxRate, taxDeductions)

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

func calculate(deduct float64, rate TaxRate, taxDeducts []TaxDeduction) float64 {

	baseTax := 150000.0
	for _, v := range taxDeducts {
		if v.DefaultAmount != 0 {
			deduct -= v.MaxDeductionAmount
		}
	}
	taxfund := (deduct - baseTax) * (rate.TaxRate / 100)
	return taxfund
}
