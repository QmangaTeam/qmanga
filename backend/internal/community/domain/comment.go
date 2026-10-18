package domain

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid community content")
var ErrForbidden = errors.New("community access denied")
var ErrNotFound = errors.New("community target not found")
var ErrLimited = errors.New("community rate limited")

type Comment struct {
	ID, UserID, TitleID, ReleaseID, ParentID, RootID, Body string
	Spoiler, Hidden                                        bool
}

func ValidBody(body string) bool {
	return utf8.ValidString(body) && utf8.RuneCountInString(strings.TrimSpace(body)) >= 1 && utf8.RuneCountInString(body) <= 2000
}
func (comment *Comment) Edit(actor, body string, spoiler bool) error {
	if actor != comment.UserID || comment.Hidden {
		return ErrForbidden
	}
	if !ValidBody(body) {
		return ErrInvalid
	}
	comment.Body = strings.TrimSpace(body)
	comment.Spoiler = spoiler
	return nil
}
func CanReply(title, release, parentTitle, parentRelease string) bool {
	return title == parentTitle && release == parentRelease
}
func ValidCategory(value string) bool {
	switch value {
	case "image", "translation", "spoiler", "abuse", "rights", "other":
		return true
	}
	return false
}
func ValidReason(value string) bool {
	return utf8.RuneCountInString(strings.TrimSpace(value)) >= 8 && utf8.RuneCountInString(value) <= 2000
}
