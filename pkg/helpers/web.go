package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/elkoshar/bookcabin/pkg/validator"
)

func ParseBodyAndValidate(r *http.Request, req interface{}) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return err
	}

	_, err = validator.ValidateStruct(req)
	if err != nil {
		return err
	}

	return nil
}
