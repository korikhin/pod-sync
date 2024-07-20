package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Statuses
const (
	statusOK    = "ok"
	statusError = "error"
)

var (
	ErrInternal       = Error("internal server error")
	ErrBadRequest     = Error("bad request")
	ErrClientNotFound = Error("no such client")
	ErrStatusNotFound = Error("no such status")
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func OK(m string) Response {
	return Response{
		Status:  statusOK,
		Message: m,
	}
}

func Error(m string) Response {
	return Response{
		Status:  statusError,
		Message: m,
	}
}

type Client struct {
	Name     *string  `json:"name" validate:"required"`
	Version  *int     `json:"version" validate:"required"`
	Image    *string  `json:"image" validate:"required"`
	CPU      *string  `json:"cpu" validate:"required"`
	Memory   *string  `json:"mem" validate:"required"`
	Priority *float64 `json:"priority" validate:"required"`
}

type Status struct {
	VWAP *bool `json:"VWAP" validate:"required"`
	TWAP *bool `json:"TWAP" validate:"required"`
	HFT  *bool `json:"HFT" validate:"required"`
}

func Validate(p interface{}) error {
	if err := validator.New().Struct(p); err != nil {
		errs := err.(validator.ValidationErrors)
		return formatErrors(errs)
	}

	return nil
}

func formatErrors(errs validator.ValidationErrors) error {
	var messages []string
	for _, err := range errs {
		var message string
		switch f := err.Field(); err.ActualTag() {
		case "required":
			message = fmt.Sprintf("field %s is required", f)
		default:
			message = fmt.Sprintf("field %s is not valid", f)
		}
		messages = append(messages, message)
	}

	return errors.New(strings.Join(messages, ", "))
}
