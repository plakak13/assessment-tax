package helper

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type CustomValidateStruct struct {
	InvalidField int `validate:"required"`
}

func TestCustomValidate(t *testing.T) {
	t.Run("Validate Success", func(t *testing.T) {

		cusVa := CustomValidator{validator: validator.New()}
		inva := CustomValidateStruct{InvalidField: 1}

		err := cusVa.Validate(inva)
		assert.NoError(t, err)
		assert.Nil(t, err)
	})

	t.Run("Validate Failed", func(t *testing.T) {
		cusVa := CustomValidator{validator: validator.New()}
		inva := CustomValidateStruct{}
		err := cusVa.Validate(inva)
		assert.Error(t, err)
	})
}

func TestNewCustomValidate(t *testing.T) {
	cv := NewValidator()

	assert.NotNil(t, cv)
	assert.NotNil(t, cv.validator)
	assert.IsType(t, validator.New(), cv.validator)
}
