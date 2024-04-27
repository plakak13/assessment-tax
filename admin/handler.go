package admin

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/plakak13/assessment-tax/helper"
	"github.com/plakak13/assessment-tax/postgres"
	"github.com/plakak13/assessment-tax/tax"
)

type Handler struct {
	store Storer
}

type Storer interface {
	UpdateTaxDeduction(s postgres.SettingTaxDeduction) (sql.Result, error)
	TaxDeductionByType(allowanceTypes []string) ([]tax.TaxDeduction, error)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func (h *Handler) AdminHandler(c echo.Context) error {
	payload := new(Setting)
	param := c.Param("type")

	if err := c.Bind(payload); err != nil {
		return helper.FailedHandler(c, "Invalid JSON", http.StatusBadRequest)
	}

	err := c.Validate(payload)

	if err != nil {
		return helper.FailedHandler(c, err.Error(), http.StatusBadRequest)
	}

	tRows, err := h.store.TaxDeductionByType([]string{param})
	if err != nil {
		return helper.FailedHandler(c, err.Error())
	}

	if tRows[0].AdminOverrideMax < payload.Amount {
		msg := fmt.Sprintf("ยอดที่กำหนดมีค่าเกินกว่า (%.1f) ที่สามารถกำหนดได้", tRows[0].AdminOverrideMax)
		return helper.FailedHandler(c, msg, http.StatusBadRequest)
	}

	if tRows[0].MinAmount > payload.Amount {
		msg := fmt.Sprintf("กรุณากำหนดมีค่าเกินกว่า (%.1f)", tRows[0].MinAmount)
		return helper.FailedHandler(c, msg, http.StatusBadRequest)
	}

	s := postgres.SettingTaxDeduction{
		ID:     tRows[0].ID,
		Amount: payload.Amount,
	}

	row, err := h.store.UpdateTaxDeduction(s)
	if err != nil {
		return helper.FailedHandler(c, err.Error())
	}
	_, err = row.RowsAffected()
	if err != nil {
		return helper.FailedHandler(c, err.Error())
	}

	return helper.SuccessHandler(c, SettingResponse{PersonalDeduction: payload.Amount})
}
