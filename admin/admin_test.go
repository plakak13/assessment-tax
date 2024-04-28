package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/plakak13/assessment-tax/helper"
	"github.com/plakak13/assessment-tax/postgres"
	"github.com/plakak13/assessment-tax/tax"
	"github.com/stretchr/testify/assert"
)

type MockAdmin struct {
	taxDeductions     []tax.TaxDeduction
	sqlResult         sql.Result
	updateError       error
	taxDeductionError error
	errorExpected     error
}

func (m MockAdmin) UpdateTaxDeduction(s postgres.SettingTaxDeduction) (sql.Result, error) {
	if m.updateError != nil {
		return nil, m.updateError
	}
	return m.sqlResult, nil
}

func (m MockAdmin) TaxDeductionByType([]string) ([]tax.TaxDeduction, error) {
	if m.taxDeductionError != nil {
		return nil, m.taxDeductionError
	}
	return m.taxDeductions, nil
}

func TestAdminHandler_Personal_Type_Success(t *testing.T) {
	e := echo.New()
	e.Validator = helper.NewValidator()

	body := `{"amount":60000}`
	req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.SetPath("/admin/deduction/:type")
	c.SetParamNames("type")
	c.SetParamValues("personal")

	h := New(MockAdmin{
		taxDeductions: []tax.TaxDeduction{
			{
				ID:                 3,
				MaxDeductionAmount: 100000,
				DefaultAmount:      60000,
				AdminOverrideMax:   100000,
				MinAmount:          10000,
				TaxAllowanceType:   "personal",
			},
		},
		sqlResult: sqlmock.NewResult(0, 1),
	})

	err := h.AdminHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminHandler_K_Receipt_Type_Success(t *testing.T) {
	e := echo.New()
	e.Validator = helper.NewValidator()

	body := `{"amount":60000}`
	req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.SetPath("/admin/deduction/:type")
	c.SetParamNames("type")
	c.SetParamValues("k-receipt")

	h := New(MockAdmin{
		taxDeductions: []tax.TaxDeduction{
			{
				ID:                 2,
				MaxDeductionAmount: 100000,
				DefaultAmount:      50000,
				AdminOverrideMax:   100000,
				MinAmount:          0,
				TaxAllowanceType:   "k-receipt",
			},
		},
		sqlResult: sqlmock.NewResult(0, 1),
	})

	err := h.AdminHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminHandler_Failed(t *testing.T) {
	t.Run("Invalid setting Type", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `{"amount":60000}`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("pesernal")

		h := New(MockAdmin{
			taxDeductionError: errors.New("invalid input value for enum tax_allowance_type"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "invalid input value for enum tax_allowance_type", jsonMashal(rec.Body.Bytes()).Message)
	})
	t.Run("Invalid JSON", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `Invalid JSON`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		h := New(MockAdmin{
			errorExpected: errors.New("Invalid JSON"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "Invalid JSON", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("Amount for setting is 0", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `{"amont": 0}`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		h := New(MockAdmin{
			errorExpected: errors.New("Field Amount is required"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "Field Amount is required", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("Amount for setting is less than default", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `{"amount": 100}`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		h := New(MockAdmin{
			taxDeductions: []tax.TaxDeduction{
				{
					ID:                 3,
					MaxDeductionAmount: 100000,
					DefaultAmount:      60000,
					AdminOverrideMax:   100000,
					MinAmount:          10000,
					TaxAllowanceType:   "personal",
				},
			},
			errorExpected: errors.New("กรุณากำหนดมีค่าเกินกว่า (10000.0)"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "กรุณากำหนดมีค่าเกินกว่า (10000.0)", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("Amount for setting is more than max override", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `{"amount": 1000000}`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		h := New(MockAdmin{
			taxDeductions: []tax.TaxDeduction{
				{
					ID:                 3,
					MaxDeductionAmount: 100000,
					DefaultAmount:      60000,
					AdminOverrideMax:   100000,
					MinAmount:          10000,
					TaxAllowanceType:   "personal",
				},
			},
			errorExpected: errors.New("ยอดที่กำหนดมีค่าเกินกว่า (100000.0) ที่สามารถกำหนดได้"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "ยอดที่กำหนดมีค่าเกินกว่า (100000.0) ที่สามารถกำหนดได้", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("Tax deduction error", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `{"amount": 50000}`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		h := New(MockAdmin{
			taxDeductions: []tax.TaxDeduction{
				{
					ID:                 3,
					MaxDeductionAmount: 100000,
					DefaultAmount:      60000,
					AdminOverrideMax:   100000,
					MinAmount:          10000,
					TaxAllowanceType:   "personal",
				},
			},
			taxDeductionError: errors.New("get tax deduction failed"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "get tax deduction failed", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("Update error and scan row failed", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `{"amount": 50000}`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		h := New(MockAdmin{
			taxDeductions: []tax.TaxDeduction{
				{
					ID:                 3,
					MaxDeductionAmount: 100000,
					DefaultAmount:      60000,
					AdminOverrideMax:   100000,
					MinAmount:          10000,
					TaxAllowanceType:   "personal",
				},
			},
			sqlResult:     sqlmock.NewErrorResult(errors.New("row affected error")),
			errorExpected: errors.New("row affected error"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "row affected error", jsonMashal(rec.Body.Bytes()).Message)
	})

	t.Run("Update amount error", func(t *testing.T) {
		e := echo.New()
		e.Validator = helper.NewValidator()

		body := `{"amount": 50000}`
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/:type", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/admin/deduction/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		h := New(MockAdmin{
			taxDeductions: []tax.TaxDeduction{
				{
					ID:                 3,
					MaxDeductionAmount: 100000,
					DefaultAmount:      60000,
					AdminOverrideMax:   100000,
					MinAmount:          10000,
					TaxAllowanceType:   "personal",
				},
			},
			updateError: errors.New("Update tax_deduction failed"),
		})

		err := h.AdminHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "Update tax_deduction failed", jsonMashal(rec.Body.Bytes()).Message)
	})
}

func jsonMashal(b []byte) helper.ErrorMessage {
	var eMsg helper.ErrorMessage
	json.Unmarshal(b, &eMsg)
	return eMsg
}
