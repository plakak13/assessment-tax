package tax

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockTax struct {
	taxRate           TaxRate
	taxRates          []TaxRate
	taxDeductions     []TaxDeduction
	errorTaxDeduction error
	errorTaxRate      error
	errorBindContext  error
	errorCalculateCSV error
}

func (h MockTax) TaxDeductionByType(allowanceTypes []string) ([]TaxDeduction, error) {
	if h.errorTaxDeduction != nil {
		return nil, h.errorTaxDeduction
	}
	return h.taxDeductions, nil
}

func (h MockTax) TaxRatesIncome(finalIncome float64) (TaxRate, error) {
	if h.errorTaxRate != nil {
		return h.taxRate, h.errorTaxRate
	}
	return h.taxRate, nil
}

func (h MockTax) CalculationHandler(echo.Context) error {
	if h.errorBindContext != nil {
		return h.errorBindContext
	}
	return nil
}

func (h MockTax) CalculationCSV() error {
	if h.errorCalculateCSV != nil {
		return h.errorCalculateCSV
	}
	return nil
}

func (h MockTax) TaxRates() ([]TaxRate, error) {
	if h.errorTaxRate != nil {
		return nil, h.errorTaxRate
	}
	return h.taxRates, nil
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

	h := New(MockTax{
		taxRates: []TaxRate{
			{
				ID:               1,
				LowerBoundIncome: 0.0,
				TaxRate:          0,
			},
			{
				ID:               2,
				LowerBoundIncome: 150001.0,
				TaxRate:          10,
			},
			{
				ID:               3,
				LowerBoundIncome: 500001.0,
				TaxRate:          15,
			},
		},
	})
	err := h.CalculationHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

}

func TestCalculationHandler_BadRequest(t *testing.T) {
	t.Run("tax with holding is 0 should retrun bad request", func(t *testing.T) {
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

		h := New(&MockTax{
			taxRates: []TaxRate{
				{
					ID:               1,
					LowerBoundIncome: 0.0,
					TaxRate:          0,
				},
				{
					ID:               2,
					LowerBoundIncome: 150001.0,
					TaxRate:          10,
				},
				{
					ID:               3,
					LowerBoundIncome: 500001.0,
					TaxRate:          15,
				},
			},
		})
		err := h.CalculationHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "\"invalid withholding tax amount\"\n", rec.Body.String())
	})

	t.Run("tax with holding over than total income should retrun badRequest", func(t *testing.T) {
		e := echo.New()
		body := `{
				"totalIncome": 500000.0,
				"wht": 500001.0,
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
	})

	t.Run("worng json body should retrun bad request", func(t *testing.T) {
		e := echo.New()
		body := "{"

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := New(&MockTax{
			errorBindContext: errors.New("error bind"),
		})
		err := h.CalculationHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Tax Deduction By Type Error should retrun internal error", func(t *testing.T) {
		e := echo.New()
		body := `{
			"totalIncome": 500000.0,
			"wht": 500001.0,
			"allowances": [
				{"allowanceType": "doantion","amount": 0.0}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := New(&MockTax{
			errorTaxDeduction: errors.New("error tax deduction by type"),
		})
		err := h.CalculationHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "\"error tax deduction by type\"\n", rec.Body.String())
	})

	t.Run("Tax Rate Error should return internal error", func(t *testing.T) {
		e := echo.New()
		body := `{
				"totalIncome": 500000.0,
				"wht": 1000.0,
				"allowances": [
					{"allowanceType": "donation","amount": 0.0}
				]
			}`
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := New(&MockTax{
			errorTaxRate: errors.New("error tax rate"),
		})
		err := h.CalculationHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "\"error tax rate\"\n", rec.Body.String())
	})
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
		errorExpectMsg string
		taxDeductions  []TaxDeduction
		taxCalculation TaxCalculation
	}
	tests := []test{
		{
			name:           "with holding tax is 0.0",
			errorExpectMsg: "invalid withholding tax amount",
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
			name:           "with holding tax is 25000.0",
			errorExpectMsg: "",
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

		if got != nil {
			if got.Error() != v.errorExpectMsg {
				t.Errorf("Expect %v but got %v", v.errorExpectMsg, got)
			}
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

func TestCalculationCSV_Success(t *testing.T) {

	e := echo.New()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("totalIncome,wht,donation\n10000,1000,500"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/calculation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := New(&MockTax{
		taxRates: []TaxRate{
			{
				ID:               1,
				LowerBoundIncome: 0.0,
				TaxRate:          0,
			},
			{
				ID:               2,
				LowerBoundIncome: 150001.0,
				TaxRate:          10,
			},
			{
				ID:               3,
				LowerBoundIncome: 500001.0,
				TaxRate:          15,
			},
		},
		taxDeductions: []TaxDeduction{
			{MaxDeductionAmount: 60000, TaxAllowanceType: "personal"},
			{MaxDeductionAmount: 100000, TaxAllowanceType: "donation"},
		},
	})
	if err := h.CalculationCSV(c); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rec.Code)
	}

}

func TestCalculationCSV_Failure(t *testing.T) {

	e := echo.New()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte(`InvalidCSVData`))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/calculation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := New(&MockTax{
		errorCalculateCSV: errors.New("error tax deduction by type"),
	})

	h.CalculationCSV(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status BadRequest, got %v", rec.Code)
	}
}
