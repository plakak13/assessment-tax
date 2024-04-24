package postgres

import (
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

		got := p.UpdateTaxDeduction(SettingTaxDeduction{ID: 1, Amount: 70000})
		rf, err := got.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, 1, int(rf))
	})
}
