package postgres

import (
	"fmt"
	"time"

	"github.com/plakak13/assessment-tax/tax"
)

type TaxDeduction struct {
	ID                 int        `postgres:"id"`
	MaxDeductionAmount float64    `postgres:"max_deduction_amount"`
	DefaultAmount      float64    `postgres:"default_amount"`
	AdminOverrideMax   float64    `postgres:"admin_override_max"`
	MinAmount          float64    `postgres:"min_amount"`
	TaxAllowanceType   string     `postgres:"tax_allowance_type"`
	CreatedAt          time.Time  `postgres:"created_at"`
	UpdatedAt          *time.Time `postgres:"updated_at"`
}

type TaxRate struct {
	ID               int        `postgres:"id"`
	LowerBoundIncome float64    `postgres:"lower_bound_income"`
	TaxRate          float64    `postgres:"tax_rate"`
	CreatedAt        time.Time  `postgres:"created_at"`
	UpdatedAt        *time.Time `postgres:"updated_at"`
}

func (p *Postgres) TaxDeductionByType(allowanceTypes []string) ([]tax.TaxDeduction, error) {

	if len(allowanceTypes) == 0 {
		return []tax.TaxDeduction{}, fmt.Errorf("please sent allowance type as least 1")
	}

	var td []tax.TaxDeduction
	argsTax := make([]any, len(allowanceTypes))

	query := "SELECT id, max_deduction_amount, default_amount, admin_override_max, min_amount, tax_allowance_type FROM tax_deduction WHERE tax_allowance_type IN ("

	for i, att := range allowanceTypes {
		query += fmt.Sprintf("$%d", i+1)
		argsTax[i] = att
		if i < len(allowanceTypes)-1 {
			query += ", "
		}
	}
	query += ")"

	stmt, err := p.Db.Prepare(query)
	if err != nil {
		return td, err
	}

	rows, err := stmt.Query(argsTax...)
	if err != nil {
		return td, err
	}

	defer rows.Close()

	for rows.Next() {
		var t tax.TaxDeduction
		err = rows.Scan(
			&t.ID,
			&t.MaxDeductionAmount,
			&t.DefaultAmount,
			&t.AdminOverrideMax,
			&t.MinAmount,
			&t.TaxAllowanceType,
		)
		if err != nil {
			return td, err
		}
		td = append(td, tax.TaxDeduction{
			ID:                 t.ID,
			MaxDeductionAmount: t.MaxDeductionAmount,
			DefaultAmount:      t.DefaultAmount,
			AdminOverrideMax:   t.AdminOverrideMax,
			MinAmount:          t.MinAmount,
			TaxAllowanceType:   t.TaxAllowanceType,
		})
	}

	return td, nil
}

func (p *Postgres) TaxRates() ([]tax.TaxRate, error) {
	query := `SELECT id, lower_bound_income, tax_rate FROM tax_rate`
	rows, err := p.Db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var tr []tax.TaxRate

	for rows.Next() {
		var t tax.TaxRate
		err = rows.Scan(&t.ID, &t.LowerBoundIncome, &t.TaxRate)
		if err != nil {
			return nil, err
		}
		tr = append(tr, tax.TaxRate{
			ID:               t.ID,
			LowerBoundIncome: t.LowerBoundIncome,
			TaxRate:          t.TaxRate,
		})
	}
	return tr, nil
}
