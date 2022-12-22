package internal

import "regexp"

var (
	commentRegex = regexp.MustCompile(`(?s)<!--(.*?)-->`)
)
