package req

import (
	"net/http"
	"url/short/pkg/errors"
)

func HandleBody[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	body, err := Decode[T](r.Body)
	if err != nil {
		appErr := errors.NewAppError(errors.ErrCodeValidationFailed, "Invalid JSON format")
		errors.WriteError(w, appErr, http.StatusBadRequest)
		return nil, appErr
	}

	err = IsValid(body)
	if err != nil {
		appErr := errors.NewAppError(errors.ErrCodeValidationFailed, "Validation failed").WithDetails(err.Error())
		errors.WriteError(w, appErr, http.StatusBadRequest)
		return nil, appErr
	}

	return &body, nil
}
