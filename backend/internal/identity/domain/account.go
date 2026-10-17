package domain

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalid     = errors.New("invalid account data")
	ErrCredentials = errors.New("invalid credentials")
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("conflict")
	ErrExpired     = errors.New("expired or consumed token")
	ErrForbidden   = errors.New("forbidden")
)

type Account struct {
	ID, Email, PasswordHash, Role, DisplayName string
	Verified                                   bool
}

func NormalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) > 254 || strings.ContainsAny(value, "\r\n") {
		return "", ErrInvalid
	}
	parsed, parseError := mail.ParseAddress(value)
	if parseError != nil || parsed.Address != value || !strings.Contains(value, "@") {
		return "", ErrInvalid
	}
	return value, nil
}
func NewAccount(id, email, passwordHash string) (Account, error) {
	normalized, validationError := NormalizeEmail(email)
	if validationError != nil || id == "" || passwordHash == "" {
		return Account{}, ErrInvalid
	}
	return Account{ID: id, Email: normalized, PasswordHash: passwordHash, Role: "reader"}, nil
}
func ValidatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < 10 || length > 128 || len(password) > 512 {
		return ErrInvalid
	}
	return nil
}
func NormalizeDisplayName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > 40 || (value != "" && utf8.RuneCountInString(value) < 2) || strings.ContainsAny(value, "\r\n\t") {
		return "", ErrInvalid
	}
	return value, nil
}
