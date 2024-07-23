package api

import (
	"fmt"
	"reflect"
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
	X *bool `json:"X" validate:"required"`
	Y *bool `json:"Y" validate:"required"`
	Z *bool `json:"Z" validate:"required"`
}

func NewValidator(opts ...validator.Option) *validator.Validate {
	v := validator.New(opts...)
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

func Validate(v *validator.Validate, p interface{}) error {
	if err := v.Struct(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			return formatErrors(errs)
		}
		return err
	}

	return nil
}

func formatErrors(errs validator.ValidationErrors) error {
	if len(errs) == 0 {
		return nil
	}
	var msg string
	switch e := errs[0]; e.Tag() {
	case "required":
		msg = fmt.Sprintf("field %s is required", e.Field())
	default:
		msg = fmt.Sprintf("field %s is not valid", e.Field())
	}

	if n := len(errs); n > 1 {
		return fmt.Errorf("%s (and %d more)", msg, n-1)
	}
	return fmt.Errorf(msg)
}
