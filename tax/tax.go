package tax

type Allowance struct {
	AllowanceType string  `json:"allowanceType" example:"donation"`
	Amount        float64 `json:"amount" example:"100.00"`
}

type TaxCalculation struct {
	TotalIncome    float64     `json:"totalIncome" validate:"required" example:"1000.00"`
	WithHoldingTax float64     `json:"wht" example:"0.0"`
	Allowances     []Allowance `json:"allowances" validate:"required"`
}

type TaxDeduction struct {
	ID                 int     `json:"id" example:"1"`
	MaxDeductionAmount float64 `json:"max_deduction_amount" example:"100.00"`
	DefaultAmount      float64 `json:"default_amount" example:"100.00"`
	AdminOverrideMax   float64 `json:"admin_override_max" example:"100.00"`
	MinAmount          float64 `json:"min_amount" example:"100.00"`
	TaxAllowanceType   string  `json:"tax_allowance_type" example:"donation"`
}

type TaxRate struct {
	ID               int     `json:"id" example:"1"`
	LowerBoundIncome float64 `json:"lower_bound_income" example:"100.0"`
	TaxRate          float64 `json:"tax_rate" example:"100.0"`
}

type CalculationResponse struct {
	Tax       float64        `json:"tax" example:"100.0"`
	TaxRefund float64        `json:"taxRefund" example:"1000.0"`
	TaxLevel  []TaxLevelInfo `json:"taxLevel" example:"taxAllowance"`
}

type TaxLevelInfo struct {
	Level string  `json:"level" example:"0-150,000"`
	Tax   float64 `json:"tax" example:"100.0"`
}

type TaxCSVCalculation struct {
	Taxes []TaxWithTotalIncome
}

type TaxWithTotalIncome struct {
	TotalIncome float64 `json:"totalIncome" example:"100.0"`
	TaxAmount   float64 `json:"tax" example:"100.0"`
}
