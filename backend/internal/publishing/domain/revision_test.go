package domain

import (
	"errors"
	"testing"
)

func TestOnlyCompleteSubmittedRevisionCanPublish(testingTarget *testing.T) {
	release := Release{ID: "release", ProjectID: "project", ChapterID: "chapter"}
	revision := NewRevision("revision", release.ID)
	if release.Publish(&revision) == nil {
		testingTarget.Fatal("empty draft published")
	}
	if reserveError := revision.ReservePage(1024, 1000, 4000, "png"); reserveError != nil {
		testingTarget.Fatal(reserveError)
	}
	if revision.Submit() == nil {
		testingTarget.Fatal("unprocessed revision submitted")
	}
	if readyError := revision.PageReady(); readyError != nil {
		testingTarget.Fatal(readyError)
	}
	if submitError := revision.Submit(); submitError != nil {
		testingTarget.Fatal(submitError)
	}
	if reserveError := revision.ReservePage(512, 500, 500, "png"); !errors.Is(reserveError, ErrImmutable) {
		testingTarget.Fatalf("submitted revision mutated: %v", reserveError)
	}
	if publishError := release.Publish(&revision); publishError != nil {
		testingTarget.Fatal(publishError)
	}
	if release.ActiveRevisionID != revision.ID() {
		testingTarget.Fatal("active revision missing")
	}
	correction := NewRevision("correction", release.ID)
	if release.Publish(&correction) == nil {
		testingTarget.Fatal("incomplete correction published")
	}
	if release.ActiveRevisionID != revision.ID() {
		testingTarget.Fatal("working publication changed during a failed correction")
	}
}

func TestUploadLimitsAndSliceBounds(testingTarget *testing.T) {
	revision := NewRevision("revision", "release")
	for _, input := range []struct {
		bytes         int64
		width, height int
		format        string
	}{
		{26 << 20, 100, 100, "png"}, {1024, 10000, 10000, "png"}, {1024, 0, 100, "png"}, {1024, 100, 100, "gif"},
	} {
		if revision.ReservePage(input.bytes, input.width, input.height, input.format) == nil {
			testingTarget.Fatal("invalid upload accepted")
		}
	}
	fragments, fragmentError := PlanFragments(1000, 20000, 720)
	if fragmentError != nil {
		testingTarget.Fatal(fragmentError)
	}
	var totalHeight int
	for _, fragment := range fragments {
		if fragment.Height > 2048 || fragment.Width > 720 {
			testingTarget.Fatal("unbounded fragment")
		}
		totalHeight += fragment.Height
	}
	if totalHeight != 14400 {
		testingTarget.Fatalf("page lost pixels: %d", totalHeight)
	}
	fragments, _ = PlanFragments(300, 400, 1600)
	if fragments[0].Width != 300 {
		testingTarget.Fatal("small original was upscaled")
	}
}
