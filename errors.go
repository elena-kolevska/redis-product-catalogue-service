package main

type Error struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

var serverError = Error{
	Title:       "Server error",
	Description: "We're sorry, something went wrong on our side :(",
}

var validationError = Error{
	Title:       "Validation errors",
	Description: "Please refer to the documentation (/documentation) for the correct input data format",
}

var notFoundError = Error{
	Title:       "Not found",
	Description: "That resource doesn't exist in our database",
}
