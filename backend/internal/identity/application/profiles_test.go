package application

import (
	"errors"
	"qmanga/internal/identity/domain"
	"strings"
	"testing"
)

func TestPreferenceValidation(t *testing.T) {
	for _, p := range []Preferences{{Version: -1}, {Bio: strings.Repeat("ә", 501)}, {Bio: "bad\x00bio"}} {
		if !errors.Is(ValidatePreferences(p), domain.ErrInvalid) {
			t.Fatal("invalid preference accepted")
		}
	}
	if err := ValidatePreferences(Preferences{Bio: strings.Repeat("ә", 500)}); err != nil {
		t.Fatal(err)
	}
}
