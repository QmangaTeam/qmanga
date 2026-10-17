package domain

import "testing"

func TestRestoredRevisionRetainsImmutableRules(testingTarget *testing.T) {
	revision, failure := RestoreRevision("revision", "release", Submitted, 9, 2, 2, 1024)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if revision.ReservePage(10, 20, 30, "png") != ErrImmutable {
		testingTarget.Fatal("restored submitted revision editable")
	}
	if _, failure = RestoreRevision("revision", "release", Ready, 1, 2, 1, 1); failure == nil {
		testingTarget.Fatal("incomplete revision restored ready")
	}
}
