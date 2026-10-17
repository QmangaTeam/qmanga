package domain

import (
	"errors"
	"strings"
)

var (
	ErrImmutable    = errors.New("revision is immutable")
	ErrTransition   = errors.New("invalid publication transition")
	ErrInvalidImage = errors.New("image exceeds upload limits")
	ErrForbidden    = errors.New("publication access denied")
	ErrNotFound     = errors.New("publication not found")
	ErrConflict     = errors.New("publication version conflict")
	ErrWithdrawn    = errors.New("publication withdrawn")
)

type State string

const (
	Draft            State = "draft"
	Processing       State = "processing"
	Ready            State = "ready"
	Submitted        State = "submitted"
	Published        State = "published"
	ChangesRequested State = "changes_requested"
)

type Revision struct {
	id         string
	releaseID  string
	state      State
	version    int64
	pageCount  int
	readyCount int
	bytes      int64
}

func NewRevision(id, releaseID string) Revision {
	return Revision{id: id, releaseID: releaseID, state: Draft, version: 1}
}
func RestoreRevision(id, releaseID string, state State, version int64, pages, ready int, bytes int64) (Revision, error) {
	if id == "" || releaseID == "" || version < 1 || pages < 0 || pages > 300 || ready < 0 || ready > pages || bytes < 0 || bytes > 500<<20 {
		return Revision{}, ErrTransition
	}
	switch state {
	case Draft, Processing, Ready, Submitted, Published, ChangesRequested:
	default:
		return Revision{}, ErrTransition
	}
	if (state == Ready || state == Submitted || state == Published) && (pages == 0 || ready != pages) {
		return Revision{}, ErrTransition
	}
	return Revision{id: id, releaseID: releaseID, state: state, version: version, pageCount: pages, readyCount: ready, bytes: bytes}, nil
}
func (revision Revision) ID() string        { return revision.id }
func (revision Revision) ReleaseID() string { return revision.releaseID }
func (revision Revision) State() State      { return revision.state }
func (revision Revision) Version() int64    { return revision.version }
func (revision Revision) PageCount() int    { return revision.pageCount }
func (revision Revision) Bytes() int64      { return revision.bytes }

func (revision *Revision) ReservePage(byteSize int64, width, height int, format string) error {
	if revision.state != Draft && revision.state != Processing && revision.state != Ready {
		return ErrImmutable
	}
	if revision.id == "" || revision.releaseID == "" || byteSize <= 0 || byteSize > 25<<20 || revision.bytes+byteSize > 500<<20 || revision.pageCount >= 300 {
		return ErrInvalidImage
	}
	if !validDimensions(width, height) {
		return ErrInvalidImage
	}
	if format != "jpeg" && format != "jpg" && format != "png" && format != "webp" {
		return ErrInvalidImage
	}
	revision.bytes += byteSize
	revision.pageCount++
	revision.state = Processing
	revision.version++
	return nil
}

func (revision *Revision) PageReady() error {
	if revision.state != Processing || revision.readyCount >= revision.pageCount {
		return ErrTransition
	}
	revision.readyCount++
	if revision.readyCount == revision.pageCount {
		revision.state = Ready
	}
	revision.version++
	return nil
}

func (revision *Revision) Submit() error {
	if revision.state != Ready || revision.pageCount == 0 || revision.readyCount != revision.pageCount {
		return ErrTransition
	}
	revision.state = Submitted
	revision.version++
	return nil
}

func (revision *Revision) Reject(reason string) error {
	if revision.state != Submitted || strings.TrimSpace(reason) == "" {
		return ErrTransition
	}
	revision.state = ChangesRequested
	revision.version++
	return nil
}

type Release struct {
	ID               string
	ProjectID        string
	ChapterID        string
	ActiveRevisionID string
	Withdrawn        bool
}

func (release *Release) Publish(revision *Revision) error {
	if release.ID == "" || release.ProjectID == "" || release.ChapterID == "" || revision == nil || revision.releaseID != release.ID || revision.state != Submitted || revision.pageCount == 0 || revision.readyCount != revision.pageCount {
		return ErrTransition
	}
	revision.state = Published
	revision.version++
	release.ActiveRevisionID = revision.id
	release.Withdrawn = false
	return nil
}

func (release *Release) Withdraw(reason string) error {
	if release.ActiveRevisionID == "" || strings.TrimSpace(reason) == "" {
		return ErrTransition
	}
	release.Withdrawn = true
	return nil
}

type Fragment struct {
	Width, Height, Top int
	Offset             float64
}

func validDimensions(width, height int) bool {
	return width > 0 && height > 0 && width <= 100000 && height <= 100000 && int64(width)*int64(height) <= 80000000
}

func PlanFragments(width, height, targetWidth int) ([]Fragment, error) {
	if !validDimensions(width, height) || targetWidth <= 0 || targetWidth > 100000 {
		return nil, ErrInvalidImage
	}
	outputWidth := min(width, targetWidth)
	outputHeight := max(1, int(int64(height)*int64(outputWidth)/int64(width)))
	fragments := make([]Fragment, 0, (outputHeight+2047)/2048)
	for top := 0; top < outputHeight; top += 2048 {
		fragments = append(fragments, Fragment{Width: outputWidth, Height: min(2048, outputHeight-top), Top: top, Offset: float64(top) / float64(outputHeight)})
	}
	return fragments, nil
}
