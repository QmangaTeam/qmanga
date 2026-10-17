package domain

import (
	"errors"
	"math"
)

var ErrInvalid = errors.New("invalid reading state")
var ErrConflict = errors.New("reading version conflict")
var ErrNotFound = errors.New("reading resource not found")
var ErrForbidden = errors.New("reading access denied")

type Progress struct {
	TitleID, ReleaseID, RevisionID, PageID, DeviceID string
	Fraction                                         float64
	FurthestOrder                                    int
	Version, Sequence                                int64
}
type Event struct {
	ID, TitleID, ReleaseID, RevisionID, PageID, DeviceID string
	Fraction                                             float64
	Sequence, BaseVersion                                int64
	Completed                                            bool
}

func Advance(current Progress, event Event, chapterOrder int, lastPage bool) (Progress, error) {
	if event.ID == "" || len(event.ID) > 64 || event.DeviceID == "" || len(event.DeviceID) > 64 || event.Sequence < 1 || event.BaseVersion < 0 || event.ReleaseID == "" || event.RevisionID == "" || event.PageID == "" || math.IsNaN(event.Fraction) || math.IsInf(event.Fraction, 0) || event.Fraction < 0 || event.Fraction > 1 {
		return current, ErrInvalid
	}
	if current.Version != event.BaseVersion {
		return current, ErrConflict
	}
	next := Progress{TitleID: event.TitleID, ReleaseID: event.ReleaseID, RevisionID: event.RevisionID, PageID: event.PageID, Fraction: event.Fraction, DeviceID: event.DeviceID, Sequence: event.Sequence, Version: current.Version + 1, FurthestOrder: current.FurthestOrder}
	if event.Completed && lastPage && event.Fraction >= .95 && chapterOrder > next.FurthestOrder {
		next.FurthestOrder = chapterOrder
	}
	return next, nil
}
func ValidStatus(value string) bool {
	switch value {
	case "reading", "planned", "completed", "paused", "dropped":
		return true
	}
	return false
}
