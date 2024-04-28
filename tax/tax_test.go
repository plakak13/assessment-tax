package tax

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/plakak13/assessment-tax/helper"
	"github.com/stretchr/testify/assert"
)

type MockTax struct {
	taxRates          []TaxRate
	taxDeductions     []TaxDeduction
	errorTaxDeduction error
	errorTaxRate      error
	errorBindContext  error
	errorCalculateCSV error
	errorInvalidWHT   error
}

func (h MockTax) TaxDeductionByType(allowanceTypes []string) ([]TaxDeduction, error) {
	if h.errorTaxDeduction != nil {
		return nil, h.errorTaxDeduction
	}
	return h.taxDeductions, nil
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

	if h.errorInvalidWHT != nil {
		return h.errorInvalidWHT
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
	e.Validator = helper.NewValidator()

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
		e.Validator = helper.NewValidator()

		body := `{
				"totalIncome": 500000.0,
				"wht": -1.0,
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
		assert.Equal(t, "invalid withholding tax amount", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("tax with holding over than total income should retrun badRequest", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

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
		e.Validator = helper.NewValidator()

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
		e.Validator = helper.NewValidator()

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
		assert.Equal(t, "error tax deduction by type", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("Tax Rate Error should return internal error", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

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
		assert.Equal(t, "error tax rate", jsonMashal(rec.Body.Bytes()).Message)
	})
}

func TestCalculate(t *testing.T) {
	type test struct {
		name          string
		expected      float64
		totalIncome   float64
		wht           float64
		taxDeductions []TaxDeduction
		taxRate       TaxRate
	}
	tests := []test{
		{
			name:        "with donation",
			expected:    29000.0,
			totalIncome: 440000.0,
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
			expected:    4000.0,
			totalIncome: 440000.0,
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
			got := calculateTaxPayable(val.totalIncome, val.wht, val.taxRate)
			assert.Equal(t, val.expected, got)
		})
	}
}

func TestValidation(t *testing.T) {

	tests := []struct {
		name           string
		taxDeducts     []TaxDeduction
		taxCalculation TaxCalculation
		expectedError  error
	}{
		{
			name: "Invalid withholding tax - negative",
			taxDeducts: []TaxDeduction{
				{TaxAllowanceType: "personal", MinAmount: 10000},
			},
			taxCalculation: TaxCalculation{
				WithHoldingTax: -100,
				TotalIncome:    100000,
			},
			expectedError: errors.New("invalid withholding tax amount"),
		},
		{
			name: "Invalid withholding tax - exceeds total income",
			taxDeducts: []TaxDeduction{
				{TaxAllowanceType: "personal", MinAmount: 10000},
			},
			taxCalculation: TaxCalculation{
				WithHoldingTax: 150000,
				TotalIncome:    100000,
			},
			expectedError: errors.New("invalid withholding tax amount"),
		},
		{
			name: "Amount for allowance is below the minimum threshold",
			taxDeducts: []TaxDeduction{
				{TaxAllowanceType: "personal", MinAmount: 10000},
			},
			taxCalculation: TaxCalculation{
				WithHoldingTax: 25000,
				TotalIncome:    500000,
				Allowances: []Allowance{
					{AllowanceType: "personal", Amount: 500},
				},
			},
			expectedError: errors.New("amount for personal allowance is below the minimum threshold"),
		},
		{
			name: "No errors",
			taxDeducts: []TaxDeduction{
				{TaxAllowanceType: "personal", MinAmount: 10000},
			},
			taxCalculation: TaxCalculation{
				WithHoldingTax: 200,
				TotalIncome:    1000,
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validation(test.taxDeducts, test.taxCalculation)
			assert.Equal(t, test.expectedError, result)
		})
	}
}

func TestMaxDeduct(t *testing.T) {
	tests := []struct {
		name           string
		tds            []TaxDeduction
		alls           []Allowance
		expectedDeduct float64
	}{
		{
			name: "Personal allowance only",
			tds: []TaxDeduction{
				{TaxAllowanceType: "personal", MaxDeductionAmount: 60000},
			},
			alls:           nil,
			expectedDeduct: 60000,
		},
		{
			name: "Allowance lower than max deduction",
			tds: []TaxDeduction{
				{TaxAllowanceType: "personal", MaxDeductionAmount: 60000},
				{TaxAllowanceType: "donation", MaxDeductionAmount: 100000},
			},
			alls: []Allowance{
				{AllowanceType: "donation", Amount: 40000},
			},
			expectedDeduct: 100000,
		},
		{
			name: "Allowance higher than max deduction",
			tds: []TaxDeduction{
				{TaxAllowanceType: "personal", MaxDeductionAmount: 60000},
				{TaxAllowanceType: "donation", MaxDeductionAmount: 100000},
			},
			alls: []Allowance{
				{AllowanceType: "donation", Amount: 200000},
			},
			expectedDeduct: 160000,
		},
		{
			name:           "No allowances",
			tds:            nil,
			alls:           nil,
			expectedDeduct: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := maxDeduct(test.tds, test.alls)
			assert.Equal(t, test.expectedDeduct, result)
		})
	}

}

func TestCalculationCSV_Success(t *testing.T) {

	e := echo.New()
	e.Validator = helper.NewValidator()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("totalIncome,wht,donation\n10000,1000,500"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
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
	err := h.CalculationCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

}

func TestCalculationCSV_Failure(t *testing.T) {

	tests := []struct {
		name           string
		csvContent     string
		expectedIncome float64
		expectedWht    float64
		expectedAmount float64
		expectedErr    helper.ErrorMessage
	}{
		{
			name:           "Invalid CSV Data",
			csvContent:     "InvalidCSVData",
			expectedIncome: 0,
			expectedWht:    0,
			expectedAmount: 0,
			expectedErr:    helper.ErrorMessage{Message: "Invalid Header"},
		},
		{
			name:           "Empty totalIncone",
			csvContent:     "totalIncome,wht,donation\n ,1000,500",
			expectedIncome: 0,
			expectedWht:    0,
			expectedAmount: 0,
			expectedErr:    helper.ErrorMessage{Message: "total income can not be string or empty"},
		},
		{
			name:           "Empty with holding tax",
			csvContent:     "totalIncome,wht,donation\n 10000.0, ,500",
			expectedIncome: 0,
			expectedWht:    0,
			expectedAmount: 0,
			expectedErr:    helper.ErrorMessage{Message: "tax with holding (twh) can not be string or empty"},
		},
		{
			name:           "invalid donation field ",
			csvContent:     "totalIncome,wht,donation\n 10000,1000,abc",
			expectedIncome: 0,
			expectedWht:    0,
			expectedAmount: 0,
			expectedErr:    helper.ErrorMessage{Message: "donation can not be string or empty"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = helper.NewValidator()

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "test.csv")
			part.Write([]byte(test.csvContent))
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := New(&MockTax{
				errorInvalidWHT: errors.New("err"),
			})

			h.CalculationCSV(c)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}

	t.Run("Failed Get Tax Deduction By Type", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.csv")
		part.Write([]byte("totalIncome,wht,donation\n 1000,1000,500"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := New(&MockTax{
			errorTaxDeduction: errors.New("Tax Deduction Failed"),
		})

		h.CalculationCSV(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "Tax Deduction Failed", jsonMashal(rec.Body.Bytes()).Message)
	})
}

func TestRemoveBOM(t *testing.T) {
	tests := []struct {
		name           string
		input          [][]string
		expectedOutput [][]string
	}{
		{
			name: "Input with BOM",
			input: [][]string{
				{"\ufefftotalIncome", "wht", "donation"},
			},
			expectedOutput: [][]string{
				{"totalIncome", "wht", "donation"},
			},
		},
		{
			name: "Input without BOM",
			input: [][]string{
				{"totalIncome", "wht", "donation"},
			},
			expectedOutput: [][]string{
				{"totalIncome", "wht", "donation"},
			},
		},
		{
			name:           "Empty input",
			input:          [][]string{},
			expectedOutput: [][]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := removeBOM(test.input)
			assert.Equal(t, test.expectedOutput, result, "unexpected result for test: %s", test.name)
		})
	}
}

func TestRefundTax(t *testing.T) {

	tests := []struct {
		name           string
		input          float64
		expectedOutput []float64
	}{
		{
			name:           "Positive input",
			input:          10000,
			expectedOutput: []float64{10000, 0},
		},
		{
			name:           "Zero input",
			input:          0,
			expectedOutput: []float64{0, 0},
		},
		{
			name:           "Nagative input",
			input:          -10000,
			expectedOutput: []float64{0, 10000},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taxRF, taxF := refundTax(test.input)

			assert.Equal(t, test.expectedOutput[0], taxF)
			assert.Equal(t, test.expectedOutput[1], taxRF)
		})
	}

}

func TestEqualSlice(t *testing.T) {
	tests := []struct {
		name         string
		sliceA       []string
		sliceB       []string
		expectedBool bool
	}{
		{
			name:         "Equal slices",
			sliceA:       []string{"a", "b", "c"},
			sliceB:       []string{"a", "b", "c"},
			expectedBool: true,
		},
		{
			name:         "Unequal slices",
			sliceA:       []string{"a", "b", "c"},
			sliceB:       []string{"a", "b", "d"},
			expectedBool: false,
		},
		{
			name:         "Different lengths",
			sliceA:       []string{"a", "b", "c"},
			sliceB:       []string{"a", "b", "c", "d"},
			expectedBool: false,
		},
		{
			name:         "Empty slices",
			sliceA:       []string{},
			sliceB:       []string{},
			expectedBool: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := equalSlice(test.sliceA, test.sliceB)
			assert.Equal(t, test.expectedBool, result)
		})
	}
}

func jsonMashal(b []byte) helper.ErrorMessage {
	var eMsg helper.ErrorMessage
	json.Unmarshal(b, &eMsg)
	return eMsg
}
