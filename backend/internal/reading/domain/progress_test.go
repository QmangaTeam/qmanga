package domain

import "testing"

func TestProgressRequiresAnExplicitConflictResolution(testingTarget *testing.T) {
	current := Progress{Version: 4, DeviceID: "phone", Sequence: 5, FurthestOrder: 20}
	event := Event{ID: "event", DeviceID: "tablet", Sequence: 1, BaseVersion: 3, ReleaseID: "release", RevisionID: "revision", PageID: "page", Fraction: .5}
	if _, failure := Advance(current, event, 10, false); failure != ErrConflict {
		testingTarget.Fatal("a stale device overwrote the current position", failure)
	}
	event.BaseVersion = 4
	updated, failure := Advance(current, event, 10, false)
	if failure != nil || updated.Version != 5 || updated.FurthestOrder != 20 {
		testingTarget.Fatal(updated, failure)
	}
	event.DeviceID = "tablet"
	event.Sequence = 0
	if _, failure = Advance(updated, event, 10, false); failure != ErrInvalid {
		testingTarget.Fatal("invalid device sequence accepted")
	}
}
