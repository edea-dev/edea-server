package util

// SPDX-License-Identifier: EUPL-1.2

import "github.com/gorilla/schema"

var (
	// FormDecoder cache for all views
	FormDecoder = schema.NewDecoder()
)
