package tax

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockTax struct {
	taxRate       TaxRate
	taxDeductions []TaxDeduction
	err           error
}

// TaxDeductionByType implements Storer.
func (h MockTax) TaxDeductionByType(allowanceTypes []string) ([]TaxDeduction, error) {
	return h.taxDeductions, h.err
}

// TaxRates implements Storer.
func (h MockTax) TaxRates(finalIncome float64) (TaxRate, error) {
	return h.taxRate, h.err
}

func TestCalculationHandler_Success(t *testing.T) {
	e := echo.New()
	body := `{
				"totalIncome": 500000.0,
				"wht": 100.0,
				"allowances": [
					{"allowanceType": "donation","amount": 0.0}
				]
			}`
	req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := New(MockTax{})
	err := h.CalculationHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

}

func TestCalculationHandler_BadRequest(t *testing.T) {
	e := echo.New()
	body := `{
				"totalIncome": 500000.0,
				"wht": 0.0,
				"allowances": [
					{"allowanceType": "donation","amount": 0.0}
				]
			}`
	req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := New(MockTax{})
	err := h.CalculationHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

}

func TestCalculate(t *testing.T) {
	type test struct {
		name          string
		expect        float64
		finalIncome   float64
		wht           float64
		taxDeductions []TaxDeduction
		taxRate       TaxRate
	}
	tests := []test{
		{
			name:        "with donation",
			expect:      29000.0,
			finalIncome: 440000.0,
			wht:         0.0,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      0.0,
					AdminOverrideMax:   0.0,
					MinAmount:          0.0,
					TaxAllowanceType:   "donation",
				},
			},
			taxRate: TaxRate{
				LowerBoundIncome: 150001,
				TaxRate:          10,
			},
		},
		{
			name:        "with k-reciept",
			expect:      4000.0,
			finalIncome: 440000.0,
			wht:         25000.0,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      50000.0,
					AdminOverrideMax:   50000.0,
					MinAmount:          0.00,
					TaxAllowanceType:   "k-reciept",
				},
			},
			taxRate: TaxRate{
				LowerBoundIncome: 150001,
				TaxRate:          10,
			},
		},
	}

	for _, val := range tests {
		t.Run(val.name, func(t *testing.T) {
			got := calculateTaxPayable(val.finalIncome, val.wht, val.taxRate)

			if got != val.expect {
				t.Errorf("Expect %.1f but got %.1f", val.expect, got)
			}
		})
	}
}

func TestValidation(t *testing.T) {

	type test struct {
		name           string
		expect         bool
		taxDeductions  []TaxDeduction
		taxCalculation TaxCalculation
	}
	tests := []test{
		{
			name:   "with holding tax is 0.0",
			expect: false,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      0.0,
					AdminOverrideMax:   0.0,
					MinAmount:          0.0,
					TaxAllowanceType:   "donation",
				},
			},
			taxCalculation: TaxCalculation{
				TotalIncome:    500000.0,
				WithHoldingTax: 0.0,
				Allowances: []Allowance{
					{
						AllowanceType: "donation",
						Amount:        0.0,
					},
				},
			},
		},
		{
			name:   "with holding tax is 25000.0",
			expect: true,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      0.0,
					AdminOverrideMax:   0.0,
					MinAmount:          0.0,
					TaxAllowanceType:   "donation",
				},
			},
			taxCalculation: TaxCalculation{
				TotalIncome:    500000.0,
				WithHoldingTax: 25000.0,
				Allowances: []Allowance{
					{
						AllowanceType: "donation",
						Amount:        0.0,
					},
				},
			},
		},
	}

	for _, v := range tests {
		got := validation(v.taxDeductions, v.taxCalculation)

		if got != v.expect {
			t.Errorf("Expect %v but got %v", v.expect, got)
		}
	}
}

func TestMaxDeduct(t *testing.T) {
	tds := []TaxDeduction{
		{MaxDeductionAmount: 60000, TaxAllowanceType: "personal"},
		{MaxDeductionAmount: 0, TaxAllowanceType: "donation"},
	}
	alls := []Allowance{
		{AllowanceType: "donation", Amount: 0},
	}

	got := maxDeduct(tds, alls)

	expected := 60000.0

	if got != expected {
		t.Errorf("maxDeduct result = %v; want %v", got, expected)
	}
}
