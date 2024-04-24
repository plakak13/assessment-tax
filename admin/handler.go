package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	store Storer
}

type Storer interface {
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func (h *Handler) AdminHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, SettingResponse{})
}
