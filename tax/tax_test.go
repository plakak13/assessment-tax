package tax

import "testing"

func TestCalculate(t *testing.T) {
	type test struct {
		name          string
		expect        float64
		finalIncome   float64
		taxDeductions []TaxDeduction
		taxRate       TaxRate
	}
	tests := []test{
		{
			name:        "with donation",
			expect:      29000.0,
			finalIncome: 440000.0,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      0.0,
					AdminOverrideMax:   0.0,
					MinAmount:          0.0,
					TaxAllowanceType:   "donation",
				},
			},
			taxRate: TaxRate{
				LowerBoundIncome: 150001,
				TaxRate:          10,
			},
		},
		{
			name:        "with k-reciept",
			expect:      39000.0,
			finalIncome: 440000.0,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      50000.0,
					AdminOverrideMax:   50000.0,
					MinAmount:          0.00,
					TaxAllowanceType:   "k-reciept",
				},
			},
			taxRate: TaxRate{
				LowerBoundIncome: 150001,
				TaxRate:          10,
			},
		},
	}

	for _, val := range tests {
		t.Run(val.name, func(t *testing.T) {
			got := calculate(val.finalIncome, val.taxRate, val.taxDeductions)

			if got != val.expect {
				t.Errorf("Expect %.1f but got %.1f", val.expect, got)
			}
		})
	}
}

func TestValidation(t *testing.T) {

	type test struct {
		name           string
		expect         bool
		taxDeductions  []TaxDeduction
		taxCalculation TaxCalculation
	}
	tests := []test{
		{
			name:   "with holding tax is 0.0",
			expect: false,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      0.0,
					AdminOverrideMax:   0.0,
					MinAmount:          0.0,
					TaxAllowanceType:   "donation",
				},
			},
			taxCalculation: TaxCalculation{
				TotalIncome:    500000.0,
				WithHoldingTax: 0.0,
				Allowances: []Allowance{
					{
						AllowanceType: "donation",
						Amount:        0.0,
					},
				},
			},
		},
		{
			name:   "with holding tax is 7000.0",
			expect: true,
			taxDeductions: []TaxDeduction{
				{
					MaxDeductionAmount: 100000.0,
					DefaultAmount:      0.0,
					AdminOverrideMax:   0.0,
					MinAmount:          0.0,
					TaxAllowanceType:   "donation",
				},
			},
			taxCalculation: TaxCalculation{
				TotalIncome:    500000.0,
				WithHoldingTax: 7000.0,
				Allowances: []Allowance{
					{
						AllowanceType: "donation",
						Amount:        0.0,
					},
				},
			},
		},
	}

	for _, v := range tests {
		got := validation(v.taxDeductions, v.taxCalculation)

		if got != v.expect {
			t.Errorf("Expect %v but got %v", v.expect, got)
		}
	}

}
