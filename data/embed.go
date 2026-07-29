package data

import _ "embed"

// DefaultTerms is the shipped term pool compiled into the binary.
//
//go:embed terms.txt
var DefaultTerms []byte
