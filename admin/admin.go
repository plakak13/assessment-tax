package admin

type Setting struct {
	Amount float64 `json:"amount" validate:"required" example:"100.00"`
}
