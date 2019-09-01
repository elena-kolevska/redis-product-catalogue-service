package main

type ApiError struct {
	HttpStatus  int    `json:"-"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Make our struct implement the error interface
func (e *ApiError) Error() string {
	return e.Description
}

var serverError = ApiError{
	HttpStatus:  500,
	Title:       "Server error",
	Description: "We're sorry, something went wrong on our side :(",
}

var validationError = ApiError{
	HttpStatus:  422,
	Title:       "Validation errors",
	Description: "Please refer to the documentation (" + config.BaseUri + "/documentation) for the correct input data format",
}

var urlParamError = ApiError{
	HttpStatus:  400,
	Title:       "Wrong parameters",
	Description: "The id in the url needs to be a valid number",
}

var notFoundError = ApiError{
	HttpStatus:  404,
	Title:       "Not found",
	Description: "That resource doesn't exist in our database",
}
