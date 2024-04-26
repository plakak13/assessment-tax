package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/plakak13/assessment-tax/admin"
	"github.com/plakak13/assessment-tax/postgres"
	"github.com/plakak13/assessment-tax/tax"
)

func main() {

	p, err := postgres.New()
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	handler := tax.New(p)
	adminHandler := admin.New(p)

	g := e.Group("/tax")

	g.POST("/calculations", handler.CalculationHandler)
	g.POST("/calculations/upload-csv", handler.CalculationCSV)

	a := e.Group("/admin")
	a.Use(middleware.BasicAuth(authenticate))

	a.POST("/deductions/:type", adminHandler.AdminHandler)

	e.Logger.Fatal(e.Start(":1323"))

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		fmt.Println("shutting down the server")
		os.Exit(0)
	}()
}

func authenticate(username, password string, c echo.Context) (bool, error) {
	if username == os.Getenv("ADMIN_USERNAME") && password == os.Getenv("ADMIN_PASSWORD") {
		return true, nil
	}
	return false, nil
}
