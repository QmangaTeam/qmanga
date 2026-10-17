package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type Credentials struct {
	slots chan struct{}
	dummy string
}

func New() *Credentials {
	provider := &Credentials{slots: make(chan struct{}, 2)}
	provider.dummy, _ = provider.Hash("timing-only-dummy-account")
	return provider
}
func (provider *Credentials) Hash(password string) (string, error) {
	provider.slots <- struct{}{}
	defer func() { <-provider.slots }()
	salt := make([]byte, 16)
	if _, readError := rand.Read(salt); readError != nil {
		return "", readError
	}
	digest := argon2.IDKey([]byte(password), salt, 2, 64*1024, 1, 32)
	return "$argon2id$v=19$m=65536,t=2,p=1$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(digest), nil
}
func (provider *Credentials) Verify(password, encoded string) bool {
	missing := encoded == ""
	if encoded == "" {
		encoded = provider.dummy
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=65536,t=2,p=1" {
		return false
	}
	salt, saltError := base64.RawStdEncoding.DecodeString(parts[4])
	expected, hashError := base64.RawStdEncoding.DecodeString(parts[5])
	if saltError != nil || hashError != nil || len(salt) != 16 || len(expected) != 32 || len(password) > 512 {
		return false
	}
	provider.slots <- struct{}{}
	defer func() { <-provider.slots }()
	actual := argon2.IDKey([]byte(password), salt, 2, 64*1024, 1, 32)
	return subtle.ConstantTimeCompare(actual, expected) == 1 && !missing
}
func (provider *Credentials) RandomToken() (string, error) {
	random := make([]byte, 32)
	if _, readError := rand.Read(random); readError != nil {
		return "", fmt.Errorf("random token: %w", readError)
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}
func (provider *Credentials) Digest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
func (provider *Credentials) NewID() string { return uuid.NewString() }
