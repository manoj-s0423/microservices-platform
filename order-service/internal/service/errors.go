package service

import "errors"

var (
	ErrValidation          = errors.New("validation failed")
	ErrUserNotFound        = errors.New("user not found")
	ErrProductNotFound     = errors.New("product not found")
	ErrProductUnavailable  = errors.New("product is out of stock or inactive")
	ErrOrderNotFound       = errors.New("order not found")
	ErrUpstreamUnavailable = errors.New("a dependency is temporarily unavailable")
	ErrPaymentDeclined     = errors.New("payment was declined")
)
