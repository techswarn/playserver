package models

import (
	"github.com/go-playground/validator/v10"
)

//Basic struct for deployment queue implementation
type DeployRequest struct {
	Name string `json:"name" validate:"required"`
	Image string `json:"image" validate:"required"`
}

// ValidateStruct performs struct based validation
func (DeployRequest DeployRequest) ValidateStruct() []*ErrorResponse {
	// create a variable to store validation errors
	var errors []*ErrorResponse

	// create a new validator
	validate := validator.New()

	// validate the struct
	err := validate.Struct(DeployRequest)

	// if the validation is failed
    // insert the error inside "errors" variable

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.ErrorMessage = getErrorMessage(err)
			element.Field = err.Field()
			errors = append(errors, &element)
		}
	}
    // return the validation errors
	return errors
}