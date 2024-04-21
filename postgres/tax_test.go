package postgres

import (
	"database/sql"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/plakak13/assessment-tax/tax"
)

type MockDB struct{}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return &sql.Row{}
}
func TestTaxDeductionByType(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()

	p := Postgres{Db: db}

	expectedRec := []TaxDeduction{
		{
			MaxDeductionAmount: 100000.00,
			DefaultAmount:      0.00,
			AdminOverrideMax:   0.00,
			MinAmount:          0.00,
			TaxAllowanceType:   "donation",
		},
		{
			MaxDeductionAmount: 50000.00,
			DefaultAmount:      50000.00,
			AdminOverrideMax:   100000.00,
			MinAmount:          0.00,
			TaxAllowanceType:   "k-reciept",
		},
	}
	allownceType := []string{"donation", "k-reciept"}

	mockQuery := "SELECT max_deduction_amount, default_amount, admin_override_max, min_amount, tax_allowance_type FROM tax_deduction WHERE tax_allowance_type IN"
	rows := sqlmock.NewRows([]string{"max_deduction_amount", "default_amount", "admin_override_max", "min_amount", "tax_allowance_type"}).
		AddRow(100000.00, 0.00, 0.00, 0.00, "donation").
		AddRow(50000.00, 50000.00, 100000.00, 0.00, "k-reciept")

	mock.ExpectQuery(mockQuery).
		WillReturnRows(rows)

	td, err := p.TaxDeductionByType(allownceType)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(td) != len(expectedRec) {
		t.Errorf("unexpected length of tax deductions: got %d, want %d", len(td), len(expectedRec))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTaxRates(t *testing.T) {

	db, mock := NewMock()
	defer db.Close()
	postgres := Postgres{Db: db}

	expectedTaxRate := tax.TaxRate{
		LowerBoundIncome: 150001.00,
		TaxRate:          10,
	}
	rows := sqlmock.NewRows([]string{"lower_bound_income", "tax_rate"}).
		AddRow(expectedTaxRate.LowerBoundIncome, expectedTaxRate.TaxRate)

	mock.ExpectQuery("SELECT lower_bound_income, tax_rate FROM tax_rate WHERE lower_bound_income <= \\$1 ORDER BY id DESC").
		WithArgs(expectedTaxRate.LowerBoundIncome).
		WillReturnRows(rows)

	taxRate, err := postgres.TaxRates(expectedTaxRate.LowerBoundIncome)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if taxRate != expectedTaxRate {
		t.Errorf("Expected tax rate: %v, got: %v", expectedTaxRate, taxRate)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}

}

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}
