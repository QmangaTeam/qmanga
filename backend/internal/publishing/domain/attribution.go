package domain

import (
	"strings"
	"unicode/utf8"
)

type Attribution struct {
	Source, Credits string
	Confirmed       bool
}

func (input Attribution) Valid() bool {
	return utf8.RuneCountInString(strings.TrimSpace(input.Source)) >= 8 && utf8.RuneCountInString(input.Source) <= 2000 && utf8.RuneCountInString(strings.TrimSpace(input.Credits)) >= 2 && utf8.RuneCountInString(input.Credits) <= 2000 && input.Confirmed
}
