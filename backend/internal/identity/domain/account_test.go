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
