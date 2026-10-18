package domain

import "testing"

func TestAccountAndPasswordRules(testingTarget *testing.T) {
	account, accountError := NewAccount("account", "Reader@Example.com", "encoded-hash")
	if accountError != nil || account.Email != "reader@example.com" || account.Role != "reader" || account.Verified {
		testingTarget.Fatalf("bad account: %+v %v", account, accountError)
	}
	for _, email := range []string{"", "reader", "Name <reader@example.com>", "reader@example.com\r\nBcc:other@example.com"} {
		if _, createError := NewAccount("account", email, "hash"); createError == nil {
			testingTarget.Fatalf("invalid email accepted: %q", email)
		}
	}
	if ValidatePassword("short") == nil {
		testingTarget.Fatal("short password accepted")
	}
	if ValidatePassword("қазақша-құпия-123") != nil {
		testingTarget.Fatal("Unicode password rejected")
	}
}

func TestDisplayNameRules(t *testing.T) {
	for _, name := range []string{"", "Қазақ оқырманы", "Reader_42", "А-Я.1"} {
		if _, err := NormalizeDisplayName(name); err != nil {
			t.Errorf("valid nickname %q: %v", name, err)
		}
	}
	for _, name := range []string{"x", " leading", "trailing ", "two  spaces", "reader@example.com", "<script>", "line\nbreak", "emoji😀"} {
		if _, err := NormalizeDisplayName(name); err == nil {
			t.Errorf("invalid nickname accepted: %q", name)
		}
	}
}
