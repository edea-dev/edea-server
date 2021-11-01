package api

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"net/url"

	"github.com/gorilla/schema"
)

var formDecoder = schema.NewDecoder()

// ErrInvalidModelForMethod is returned in the odd case when a request doesn't match the model it should have
// this most likely means a programming error somewhere in the routes.
var ErrInvalidModelForMethod = errors.New("expected a different model type")

// API defines the methods expected from the model package to implement a generic API
type API interface {
	DecodeForm(url.Values) (interface{}, error)
	Validate(interface{}) error
	Update(interface{}, string) (interface{}, error)
	Create(interface{}, string) (interface{}, error)
	Delete(interface{}, string) error
	Get(interface{}, string) (interface{}, error)
}
