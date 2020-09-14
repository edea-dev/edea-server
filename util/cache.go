package util

import "github.com/gorilla/schema"

var (
	// FormDecoder cache for all views
	FormDecoder = schema.NewDecoder()
)
