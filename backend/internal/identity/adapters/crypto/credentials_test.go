package crypto

import "testing"

func TestHashesAndTokens(testingTarget *testing.T) {
	provider := New()
	hash, hashError := provider.Hash("correct-password-123")
	if hashError != nil {
		testingTarget.Fatal(hashError)
	}
	if !provider.Verify("correct-password-123", hash) || provider.Verify("wrong", hash) {
		testingTarget.Fatal("password verification failed")
	}
	if provider.Verify("timing-only-dummy-account", "") {
		testingTarget.Fatal("missing hash authenticated")
	}
	if provider.Verify("anything", "$argon2id$v=19$m=999999999,t=2,p=1$AAAA$AAAA") {
		testingTarget.Fatal("unbounded parameters accepted")
	}
	token, tokenError := provider.RandomToken()
	if tokenError != nil || len(token) != 43 || len(provider.Digest(token)) != 64 {
		testingTarget.Fatal("invalid token")
	}
}
