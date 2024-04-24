package tax

type Allowance struct {
	AllowanceType string  `json:"allowanceType" example:"donation"`
	Amount        float64 `json:"amount" example:"100.00"`
}

type TaxCalculation struct {
	TotalIncome    float64     `json:"totalIncome" example:"1000.00"`
	WithHoldingTax float64     `json:"wht" example:"0.0"`
	Allowances     []Allowance `json:"allowances"`
}

type TaxDeduction struct {
	MaxDeductionAmount float64 `json:"max_deduction_amount" example:"100.00"`
	DefaultAmount      float64 `json:"default_amount" example:"100.00"`
	AdminOverrideMax   float64 `json:"admin_override_max" example:"100.00"`
	MinAmount          float64 `json:"min_amount" example:"100.00"`
	TaxAllowanceType   string  `json:"tax_allowance_type" example:"donation"`
}

type TaxRate struct {
	LowerBoundIncome float64 `json:"lower_bound_income" example:"100.0"`
	TaxRate          float64 `json:"tax_rate" example:"100.0"`
}

type CalculationResponse struct {
	Tax float64 `json:"tax"`
}
