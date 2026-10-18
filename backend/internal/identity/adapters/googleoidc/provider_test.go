package googleoidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	jose "github.com/go-jose/go-jose/v4"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type localKeys struct{ key *rsa.PublicKey }

func (k localKeys) VerifySignature(_ context.Context, raw string) ([]byte, error) {
	token, err := jose.ParseSigned(raw, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return nil, err
	}
	return token.Verify(k.key)
}
func TestProviderRejectsInvalidIDTokens(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, nil)
	if err != nil {
		t.Fatal(err)
	}
	nonce := "0123456789012345678901234567890123456789012"
	for _, scenario := range []string{"valid", "issuer", "audience", "expired", "nonce", "azp", "signature"} {
		t.Run(scenario, func(t *testing.T) {
			claims := map[string]any{"iss": "https://accounts.google.com", "aud": "client", "sub": "stable", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": nonce, "email": "reader@example.com", "email_verified": true}
			switch scenario {
			case "issuer":
				claims["iss"] = "https://evil.test"
			case "audience":
				claims["aud"] = "other"
			case "expired":
				claims["exp"] = time.Now().Add(-time.Hour).Unix()
			case "nonce":
				claims["nonce"] = "different"
			case "azp":
				claims["azp"] = "other"
			}
			raw, _ := json.Marshal(claims)
			signed, _ := signer.Sign(raw)
			jwt, _ := signed.CompactSerialize()
			if scenario == "signature" {
				jwt = jwt[:len(jwt)-12] + "invalid-sign"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
				if r.Form.Get("code_verifier") != "verifier" {
					t.Error("PKCE missing")
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"access_token":"access","token_type":"Bearer","id_token":%q}`, jwt)
			}))
			defer server.Close()
			provider := New("client", "secret", "https://app.test/api/v1/auth/google/callback")
			provider.config.Endpoint.TokenURL = server.URL
			provider.verifier = oidc.NewVerifier("https://accounts.google.com", localKeys{&key.PublicKey}, &oidc.Config{ClientID: "client", SupportedSigningAlgs: []string{"RS256"}})
			identity, err := provider.Exchange(context.Background(), "code", nonce, "verifier")
			if scenario == "valid" {
				if err != nil || identity.Subject != "stable" || !identity.Verified {
					t.Fatalf("valid token: %+v %v", identity, err)
				}
			} else if err == nil {
				t.Fatalf("invalid %s token accepted", scenario)
			}
		})
	}
}
