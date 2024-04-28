package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/plakak13/assessment-tax/admin"
	"github.com/plakak13/assessment-tax/helper"
	"github.com/plakak13/assessment-tax/postgres"
	"github.com/plakak13/assessment-tax/tax"
)

func main() {

	p, err := postgres.New()
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Validator = helper.NewValidator()

	handler := tax.New(p)
	adminHandler := admin.New(p)

	g := e.Group("/tax")

	g.POST("/calculations", handler.CalculationHandler)
	g.POST("/calculations/upload-csv", handler.CalculationCSV)

	a := e.Group("/admin")
	a.Use(middleware.BasicAuth(authenticate))

	a.POST("/deductions/:type", adminHandler.AdminHandler)

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))

	go func() {
		if err := e.Start(port); err != nil && err != http.ErrServerClosed {
			e.Logger.Info("shutting down the server")
		}
	}()

	shitdown := make(chan os.Signal, 1)
	signal.Notify(shitdown, os.Interrupt, syscall.SIGTERM)
	<-shitdown
	fmt.Println("shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}

func authenticate(username, password string, c echo.Context) (bool, error) {
	if username == os.Getenv("ADMIN_USERNAME") && password == os.Getenv("ADMIN_PASSWORD") {
		return true, nil
	}
	return false, nil
}
