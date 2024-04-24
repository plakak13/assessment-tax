package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/plakak13/assessment-tax/tax"
	"github.com/stretchr/testify/assert"
)

type MockDB struct{}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return &sql.Row{}
}
func TestTaxDeductionByType(t *testing.T) {

	t.Run("Select Tax Deduction Success", func(t *testing.T) {
		db, mock := NewMock()
		defer db.Close()

		p := Postgres{Db: db}

		expectedRec := []TaxDeduction{
			{
				ID:                 1,
				MaxDeductionAmount: 100000.00,
				DefaultAmount:      0.00,
				AdminOverrideMax:   0.00,
				MinAmount:          0.00,
				TaxAllowanceType:   "donation",
			},
			{
				ID:                 2,
				MaxDeductionAmount: 50000.00,
				DefaultAmount:      50000.00,
				AdminOverrideMax:   100000.00,
				MinAmount:          0.00,
				TaxAllowanceType:   "k-reciept",
			},
		}
		allownceType := []string{"donation", "k-reciept"}

		mockQuery := "SELECT id, max_deduction_amount, default_amount, admin_override_max, min_amount, tax_allowance_type FROM tax_deduction WHERE tax_allowance_type IN ()"
		rows := sqlmock.NewRows([]string{"id", "max_deduction_amount", "default_amount", "admin_override_max", "min_amount", "tax_allowance_type"}).
			AddRow(1, 100000.00, 0.00, 0.00, 0.00, "donation").
			AddRow(2, 50000.00, 50000.00, 100000.00, 0.00, "k-reciept")

		mock.ExpectPrepare(mockQuery).
			ExpectQuery().
			WithArgs(allownceType[0], allownceType[1]).
			WillReturnRows(rows)

		td, err := p.TaxDeductionByType(allownceType)

		assert.NoError(t, err)
		assert.Equal(t, len(td), len(expectedRec), "unexpected length of tax deductions")
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, fmt.Sprintf("Unfulfilled expectations: %s", err))
	})
	t.Run("empty allwance type should return error", func(t *testing.T) {
		db, _ := NewMock()
		defer db.Close()

		p := Postgres{Db: db}
		rows, err := p.TaxDeductionByType([]string{})
		if err == nil {
			t.Errorf("expect error message but got %v", err)
		}

		if len(rows) > 0 {
			t.Errorf("expect empty record but got %v records", len(rows))
		}
	})

	t.Run("prepare query faild should return error", func(t *testing.T) {
		db, mock := NewMock()
		defer db.Close()

		p := Postgres{Db: db}
		mockQuery := "SELECT id, max_deduction_amount, default_amount, admin_override_max, min_amount, tax_allowance_type FROM tax_deduction WHERE tax_allowance_type IN ()"

		expectedError := errors.New("failed to prepare statement")
		mock.ExpectPrepare(mockQuery).WillReturnError(expectedError)

		_, err := p.TaxDeductionByType([]string{"donation", "k-reciept"})

		assert.EqualError(t, err, expectedError.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)

	})

	t.Run("query argument faild should return error", func(t *testing.T) {
		db, mock := NewMock()
		defer db.Close()

		p := Postgres{Db: db}
		mockQuery := "SELECT id, max_deduction_amount, default_amount, admin_override_max, min_amount, tax_allowance_type FROM tax_deduction WHERE tax_allowance_type IN ()"

		expectedError := errors.New("query error")

		mock.ExpectPrepare(mockQuery).ExpectQuery().WithArgs().WillReturnError(expectedError)

		_, err := p.TaxDeductionByType([]string{"donation", "k-reciept"})

		assert.EqualError(t, err, expectedError.Error())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)

	})

	t.Run("Scan faild should return error", func(t *testing.T) {
		db, mock := NewMock()
		defer db.Close()

		p := Postgres{Db: db}
		mockQuery := "SELECT id, max_deduction_amount, default_amount, admin_override_max, min_amount, tax_allowance_type FROM tax_deduction WHERE tax_allowance_type IN ()"
		rows := sqlmock.NewRows([]string{"id", "max_deduction_amount", "default_amount", "admin_override_max", "min_amount", "tax_allowance_type"}).
			AddRow(nil, nil, 50000.00, 100000.00, 0.00, "k-reciept").
			RowError(2, errors.New("error row"))

		mock.ExpectPrepare(mockQuery).
			ExpectQuery().
			WithArgs("donation", "k-reciept").
			WillReturnRows(rows)

		_, err := p.TaxDeductionByType([]string{"donation", "k-reciept"})

		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, fmt.Sprintf("Unfulfilled expectations: %s", err))

	})

}

func TestTaxRates(t *testing.T) {

	t.Run("Scan Rows Error", func(t *testing.T) {

		db, mock := NewMock()
		defer db.Close()

		expectedTaxRate := tax.TaxRate{
			LowerBoundIncome: 150001.00,
			TaxRate:          10,
		}
		allownceType := []string{"donation", "k-reciept"}
		postgres := Postgres{Db: db}
		rows := sqlmock.NewRows(allownceType).
			AddRow(nil, nil).
			RowError(1, errors.New("error rows"))

		mock.ExpectQuery("SELECT id, lower_bound_income, tax_rate FROM tax_rate WHERE lower_bound_income <= \\$1 ORDER BY id DESC").
			WithArgs(expectedTaxRate.LowerBoundIncome).
			WillReturnRows(rows)

		got, err := postgres.TaxRatesIncome(expectedTaxRate.LowerBoundIncome)

		assert.Equal(t, got, tax.TaxRate{})
		assert.Error(t, err, "should be error")

	})

	t.Run("No Rows Error", func(t *testing.T) {

		db, mock := NewMock()
		defer db.Close()

		expectedTaxRate := tax.TaxRate{
			LowerBoundIncome: 150001.00,
			TaxRate:          10,
		}

		allownceType := []string{"donation", "k-reciept"}
		postgres := Postgres{Db: db}

		rows := sqlmock.NewRows(allownceType)

		mock.ExpectQuery("SELECT id, lower_bound_income, tax_rate FROM tax_rate WHERE lower_bound_income <= \\$1 ORDER BY id DESC").
			WithArgs(expectedTaxRate.LowerBoundIncome).
			WillReturnRows(rows)

		got, err := postgres.TaxRatesIncome(expectedTaxRate.LowerBoundIncome)

		assert.Equal(t, got, tax.TaxRate{})
		assert.Error(t, err, "sql: no rows")

	})

	t.Run("Get TaxRate Success", func(t *testing.T) {
		db, mock := NewMock()
		defer db.Close()
		postgres := Postgres{Db: db}

		expectedTaxRate := tax.TaxRate{
			ID:               1,
			LowerBoundIncome: 150001.00,
			TaxRate:          101,
		}

		rows := sqlmock.NewRows([]string{"id", "lower_bound_income", "tax_rate"}).
			AddRow(expectedTaxRate.ID, expectedTaxRate.LowerBoundIncome, expectedTaxRate.TaxRate)

		mock.ExpectQuery("SELECT id, lower_bound_income, tax_rate FROM tax_rate WHERE lower_bound_income <= \\$1 ORDER BY id DESC").
			WithArgs(expectedTaxRate.LowerBoundIncome).
			WillReturnRows(rows)

		taxRate, err := postgres.TaxRatesIncome(expectedTaxRate.LowerBoundIncome)

		assert.NoError(t, err)
		assert.Equal(t, taxRate, expectedTaxRate, "they should by equal")
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, fmt.Sprintf("Unfulfilled expectations: %s", err))

	})
}

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}
