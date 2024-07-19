package storage

import "errors"

var (
	ErrMalformedConfig        = errors.New("failed to parse config")
	ErrConnectionTimeout      = errors.New("connection timed out")
	ErrConnectionDial         = errors.New("connection dial error")
	ErrConnectionInvalid      = errors.New("connection invalid")
	ErrConnectionUnauthorized = errors.New("connection unauthorized")
	ErrClientNotFound         = errors.New("client not found")
	ErrStatusNotFound         = errors.New("status not found")
)
