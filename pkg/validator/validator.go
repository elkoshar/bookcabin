package validator

import (
	"errors"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var engine *validator.Validate

// init the decoder
func init() {
	// validator.New() is safe to call multiple times since it already use sync.Pool
	engine = validator.New()
}

// ValidateStruct validate given struct that have validate tag
func ValidateStruct(object interface{}) (isValid bool, err error) {
	if reflect.ValueOf(object).Kind() == reflect.Slice ||
		reflect.ValueOf(object).Kind() == reflect.Array {
		arr := reflect.ValueOf(object)
		if arr.Len() == 0 || arr.IsZero() {
			isValid = false
			err = errors.New("object is empty")
			return
		}
		err = engine.Var(object, "required,dive")
		isValid = err == nil
		return

	}

	err = engine.Struct(object)
	isValid = err == nil
	return
}
