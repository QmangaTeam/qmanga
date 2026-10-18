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

func TestPublicTelegramLink(testingTarget *testing.T) {
	for _, value := range []string{"", "https://t.me/dala_team", "https://t.me/dala_team/"} {
		team, failure := NewTeam("id", "dala", "Дала", "", value)
		if failure != nil || (value != "" && team.TelegramURL != "https://t.me/dala_team") {
			testingTarget.Fatalf("valid Telegram link %q: %#v %v", value, team, failure)
		}
	}
	for _, value := range []string{"javascript:alert(1)", "https://evil.test/dala_team", "https://t.me@evil.test/dala_team", "https://t.me/dala_team?start=private", "https://t.me/+private", "http://t.me/dala_team"} {
		if _, failure := NewTeam("id", "dala", "Дала", "", value); failure == nil {
			testingTarget.Fatalf("unsafe link accepted: %q", value)
		}
	}
}
