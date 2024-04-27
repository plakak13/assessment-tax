package admin

type Setting struct {
	Amount float64 `json:"amount" validate:"required" example:"100.00"`
}

type SettingResponse struct {
	PersonalDeduction float64 `json:"personalDeduction" example:"1000.00"`
}
