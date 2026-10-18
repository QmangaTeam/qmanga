package domain

import "testing"

func TestCommentOwnershipAndDepth(testingTarget *testing.T) {
	comment := Comment{UserID: "owner"}
	if failure := comment.Edit("stranger", "Мәтін", false); failure != ErrForbidden {
		testingTarget.Fatal("foreign edit allowed")
	}
	if failure := comment.Edit("owner", "   ", false); failure != ErrInvalid {
		testingTarget.Fatal("empty accepted")
	}
	if !CanReply("title", "release", "title", "release") || CanReply("other", "release", "title", "release") {
		testingTarget.Fatal("invalid reply depth/context")
	}
}
