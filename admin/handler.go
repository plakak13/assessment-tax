package admin

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/plakak13/assessment-tax/postgres"
	"github.com/plakak13/assessment-tax/tax"
)

type Handler struct {
	store Storer
}

type Storer interface {
	UpdateTaxDeduction(s postgres.SettingTaxDeduction) sql.Result
	TaxDeductionByType(allowanceTypes []string) ([]tax.TaxDeduction, error)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func (h *Handler) AdminHandler(c echo.Context) error {
	payload := new(Setting)

	err := c.Bind(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	tRows, err := h.store.TaxDeductionByType([]string{"personal"})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	s := postgres.SettingTaxDeduction{
		ID:     tRows[0].ID,
		Amount: payload.Amount,
	}

	row := h.store.UpdateTaxDeduction(s)
	_, err = row.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, SettingResponse{PersonalDeduction: payload.Amount})
}
