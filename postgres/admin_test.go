package postgres

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

type MockAdmin struct {
}

func TestUpdateTaxDeduction(t *testing.T) {
	t.Run("Update Tax Deduction Success", func(t *testing.T) {
		db, mock := NewMock()
		defer db.Close()
		sqlmock.NewResult(0, 1)
		p := Postgres{Db: db}

		mockQuery := "UPDATE tax_deduction SET max_deduction_amount = \\$1 WHERE id = \\$2"
		mock.ExpectExec(mockQuery).
			WithArgs(70000.0, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		got, err := p.UpdateTaxDeduction(SettingTaxDeduction{ID: 1, Amount: 70000})

		assert.NoError(t, err)

		rf, err := got.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, 1, int(rf))

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Update Tax Deduction Failed", func(t *testing.T) {
		db, mock := NewMock()
		defer db.Close()

		p := Postgres{Db: db}

		mockQuery := "UPDATE tax_deduction SET max_deduction_amount = \\$1 WHERE id = \\$2"
		mock.ExpectExec(mockQuery).
			WithArgs(70000.0, 1).
			WillReturnError(sql.ErrNoRows)

		got, err := p.UpdateTaxDeduction(SettingTaxDeduction{ID: 1, Amount: 70000})

		assert.Error(t, err)
		assert.Nil(t, got)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
