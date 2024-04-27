package admin

import (
	"database/sql"
	"testing"

	"github.com/plakak13/assessment-tax/tax"
)

type MockAdmin struct {
	taxDeduction []tax.TaxDeduction
	sqlResult    sql.Result
	updateError  error
}

func (m *MockAdmin) UpdateTaxDeduction() sql.Result {
	if m.updateError != nil {
		return nil
	}
	return m.sqlResult
}

func TestAdminHadler(t *testing.T) {

}
