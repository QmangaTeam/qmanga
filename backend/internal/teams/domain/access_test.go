package domain

import "testing"

func TestApprovedMembershipPermissions(testingTarget *testing.T) {
	access := Access{Verified: true, Approved: true, Role: Member}
	if !access.Allows(Upload) || access.Allows(Submit) || access.Allows(Manage) {
		testingTarget.Fatal("member permissions are wrong")
	}
	access.Role = Editor
	if !access.Allows(Submit) || access.Allows(Manage) {
		testingTarget.Fatal("editor permissions are wrong")
	}
	access.Role = Owner
	if !access.Allows(Manage) {
		testingTarget.Fatal("owner cannot manage team")
	}
	access.Approved = false
	if access.Allows(Upload) || access.Allows(Submit) {
		testingTarget.Fatal("unapproved team can publish")
	}
	access.Approved = true
	access.Role = ""
	if access.Allows(ReadDraft) || access.Allows(Upload) {
		testingTarget.Fatal("revoked membership retains access")
	}
	access.GlobalRole = "moderator"
	if !access.Allows(ReadDraft) || access.Allows(Upload) {
		testingTarget.Fatal("review permission grants team editing")
	}
}
