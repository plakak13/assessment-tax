package postgres

import "database/sql"

type SettingTaxDeduction struct {
	ID     int
	Amount float64
}

type UpdateTaxDeductionResponse struct {
	RowEffectID  int
	LastInsertId int
}

func (p *Postgres) UpdateTaxDeduction(s SettingTaxDeduction) sql.Result {

	query := `UPDATE tax_deduction SET max_deduction_amount = $1 WHERE id = $2 `

	row, err := p.Db.Exec(query, s.Amount, s.ID)
	if err != nil {
		return nil
	}

	return row
}
