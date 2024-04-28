package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/plakak13/assessment-tax/admin"
	"github.com/plakak13/assessment-tax/tax"
	"github.com/stretchr/testify/assert"
)

type Response struct {
	*http.Response
	err error
}

func clientRequest(method, url string, body io.Reader, isAuthen bool) *Response {
	req, _ := http.NewRequest(method, url, body)
	if isAuthen {
		req.Header.Add("Authorization", "Basic YWRtaW5UYXg6YWRtaW4h")
	}

	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}

func (r *Response) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(v)
}

func uri(paths ...string) string {

	baseURL := "http://localhost:8080"
	return baseURL + "/" + strings.Join(paths, "/")
}

func TestCalculateTax(t *testing.T) {
	t.Run("calculate tax no allowance", func(t *testing.T) {
		calRequest := tax.TaxCalculation{
			TotalIncome:    500000,
			WithHoldingTax: 20000,
			Allowances:     []tax.Allowance{},
		}

		calculateJSON, _ := json.Marshal(calRequest)
		res := clientRequest(http.MethodPost, uri("tax", "calculations"), bytes.NewBuffer(calculateJSON), false)
		var result tax.CalculationResponse

		err := res.Decode(&result)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("calculate tax with donation and k-receipt", func(t *testing.T) {
		calRequest := tax.TaxCalculation{
			TotalIncome:    500000,
			WithHoldingTax: 20000,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "donation",
					Amount:        200000,
				},
				{
					AllowanceType: "k-receipt",
					Amount:        50000,
				},
			},
		}

		calculateJSON, _ := json.Marshal(calRequest)
		res := clientRequest(http.MethodPost, uri("tax", "calculations"), bytes.NewBuffer(calculateJSON), false)
		var result tax.CalculationResponse

		err := res.Decode(&result)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("calculate withhoding tax is zero", func(t *testing.T) {
		calRequest := tax.TaxCalculation{
			TotalIncome:    500000,
			WithHoldingTax: 0,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "donation",
					Amount:        200000,
				},
				{
					AllowanceType: "k-receipt",
					Amount:        50000,
				},
			},
		}

		calculateJSON, _ := json.Marshal(calRequest)
		res := clientRequest(http.MethodPost, uri("tax", "calculations"), bytes.NewBuffer(calculateJSON), false)

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("calculate tax withhoding more than total income", func(t *testing.T) {
		calRequest := tax.TaxCalculation{
			TotalIncome:    500000,
			WithHoldingTax: 700000,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "donation",
					Amount:        200000,
				},
				{
					AllowanceType: "k-receipt",
					Amount:        50000,
				},
			},
		}

		calculateJSON, _ := json.Marshal(calRequest)
		res := clientRequest(http.MethodPost, uri("tax", "calculations"), bytes.NewBuffer(calculateJSON), false)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestAdminHandler_Integration(t *testing.T) {
	t.Run("admin setting personal deduction success", func(t *testing.T) {
		settingRequest := admin.Setting{
			Amount: 60000,
		}

		settingJSON, _ := json.Marshal(settingRequest)
		res := clientRequest(http.MethodPost, uri("admin", "deductions", "personal"), bytes.NewBuffer(settingJSON), true)
		var result admin.SettingResponse

		err := res.Decode(&result)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

	})

	t.Run("admin setting personal deduction failed", func(t *testing.T) {
		settingRequest := admin.Setting{
			Amount: 6000,
		}

		settingJSON, _ := json.Marshal(settingRequest)
		res := clientRequest(http.MethodPost, uri("admin", "deductions", "personal"), bytes.NewBuffer(settingJSON), true)
		var result admin.SettingResponse

		err := res.Decode(&result)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	})

}
