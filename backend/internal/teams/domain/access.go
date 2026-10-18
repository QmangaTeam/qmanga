package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrForbidden = errors.New("team access denied")
	ErrInvalid   = errors.New("invalid team data")
	ErrConflict  = errors.New("team conflict")
	ErrExpired   = errors.New("invitation expired or consumed")
	ErrNotFound  = errors.New("team not found")
)

type Role string

const (
	Member Role = "member"
	Editor Role = "editor"
	Owner  Role = "owner"
)

type Action string

const (
	ReadDraft Action = "read_draft"
	Upload    Action = "upload"
	Submit    Action = "submit"
	Manage    Action = "manage"
	Approve   Action = "approve"
)

type Access struct {
	Verified, Approved bool
	Role               Role
	GlobalRole         string
}

func (access Access) Allows(action Action) bool {
	if !access.Verified {
		return false
	}
	if action == Approve {
		return access.GlobalRole == "admin"
	}
	if action == ReadDraft && (access.GlobalRole == "admin" || access.GlobalRole == "moderator") {
		return true
	}
	if access.Role != Member && access.Role != Editor && access.Role != Owner {
		return false
	}
	if action == Manage {
		return access.Role == Owner
	}
	if !access.Approved {
		return false
	}
	switch action {
	case ReadDraft, Upload:
		return true
	case Submit:
		return access.Role == Editor || access.Role == Owner
	}
	return false
}

type Team struct {
	ID, Slug, Name, Description string
	TelegramURL                 string
	ReviewStatus, ReviewNote    string
	Approved                    bool
}

var slugPattern = regexp.MustCompile("^[a-z0-9]+(-[a-z0-9]+)*$")

func NewTeam(id, slug, name, description string, telegram ...string) (Team, error) {
	slug = strings.TrimSpace(slug)
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if id == "" || len(slug) < 3 || len(slug) > 64 || !slugPattern.MatchString(slug) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 80 || strings.ContainsAny(name, "\r\n\t") || utf8.RuneCountInString(description) > 2000 {
		return Team{}, ErrInvalid
	}
	link := ""
	if len(telegram) > 0 {
		link = strings.TrimSuffix(strings.TrimSpace(telegram[0]), "/")
	}
	if link != "" && !telegramPattern.MatchString(link) {
		return Team{}, ErrInvalid
	}
	return Team{ID: id, Slug: slug, Name: name, Description: description, TelegramURL: link, ReviewStatus: "pending"}, nil
}

var telegramPattern = regexp.MustCompile(`^https://t\.me/[A-Za-z][A-Za-z0-9_]{4,31}$`)

func InvitationEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	parsed, failure := mail.ParseAddress(value)
	if failure != nil || len(value) > 254 || parsed.Address != value || strings.ContainsAny(value, "\r\n") {
		return "", ErrInvalid
	}
	return value, nil
}
func Invitable(role Role) bool { return role == Member || role == Editor }
