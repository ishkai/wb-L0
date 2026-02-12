package validation

import (
	"awesomeProject3/project/model"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateOrder(o *model.Order) error {
	return validate.Struct(o)
}
