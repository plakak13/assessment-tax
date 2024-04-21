package tax

// import (
// 	"testing"
// )

// type MockTax struct {
// 	taxrate      TaxRate
// 	taxdeduction []TaxDeduction
// 	err          error
// }

// func (s MockTax) TaxDeductionByType([]string) ([]TaxDeduction, error) {
// 	return s.taxdeduction, s.err
// }

// func (s MockTax) TaxRates(float64) (TaxRate, error) {
// 	return s.taxrate, s.err
// }

// func TestGetTaxDeduction(t *testing.T) {
// 	t.Run("dection", func(t *testing.T) {
// 		mt := MockTax{
// 			taxdeduction: []TaxDeduction{
// 				// 	{
// 				// 		MaxDeductionAmount: 100000.00,
// 				// 		DefaultAmount:      0.00,
// 				// 		AdminOverrideMax:   0.00,
// 				// 		MinAmount:          0.00,
// 				// 		TaxAllowanceType:   "donation",
// 				// 	},
// 				// 	{
// 				// 		MaxDeductionAmount: 50000.00,
// 				// 		DefaultAmount:      50000.00,
// 				// 		AdminOverrideMax:   100000.00,
// 				// 		MinAmount:          0.00,
// 				// 		TaxAllowanceType:   "k-reciept",
// 				// 	},
// 			},
// 			err: nil,
// 		}
// 		h := New(mt)

// 		td, err := h.store.TaxDeductionByType([]string{"donation", "k-reciept"})
// 		if err != nil {
// 			t.Errorf("expect tax deduction but got %v", err)
// 		}

// 		if len(td) != len(mt.taxdeduction) {
// 			t.Errorf("unexpected length of tax deductions: got %d, want %d", len(td), len(mt.taxdeduction))
// 		}

// 		emptyResult, err := h.store.TaxDeductionByType([]string{})
// 		if err != nil {
// 			t.Errorf("Unexpected error: %v", err)
// 		}
// 		if len(emptyResult) != 0 {
// 			t.Errorf("Expected empty result, got %d rows", len(emptyResult))
// 		}
// 	})
// }
