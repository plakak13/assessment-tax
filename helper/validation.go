package helper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		var msg []string
		for _, e := range err.(validator.ValidationErrors) {
			eMsg := fmt.Sprintf("Field %s is %s", e.Field(), e.ActualTag())
			msg = append(msg, eMsg)
		}
		return errors.New(strings.Join(msg, ", "))
	}
	return nil
}

func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}
