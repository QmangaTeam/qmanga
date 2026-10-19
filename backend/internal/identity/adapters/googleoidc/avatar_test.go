package googleoidc

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"qmanga/internal/identity/application"
)

type avatarUpload struct {
	calls int
	body  string
}

func (u *avatarUpload) Upload(_ context.Context, actor, kind, owner string, body io.Reader) (application.Avatar, error) {
	u.calls++
	raw, _ := io.ReadAll(body)
	u.body = string(raw)
	return application.Avatar{}, nil
}

type avatarTransport func(*http.Request) (*http.Response, error)

func (f avatarTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestGoogleAvatarRejectsUnsafeURLsAndBoundsDownload(t *testing.T) {
	for _, url := range []string{"http://lh3.googleusercontent.com/a/photo", "https://localhost/a", "https://127.0.0.1/a", "https://lh3.googleusercontent.com.evil.test/a", "https://evil.googleusercontent.com/a", "https://user@lh3.googleusercontent.com/a", "https://lh3.googleusercontent.com:444/a", "https://lh3.googleusercontent.com/a#fragment"} {
		t.Run(url, func(t *testing.T) {
			u := &avatarUpload{}
			importer := NewAvatarImporter(u)
			if err := importer.ImportGoogleAvatar(context.Background(), "user", url); err == nil || u.calls != 0 {
				t.Fatalf("unsafe URL accepted: %v", err)
			}
		})
	}
	for _, scenario := range []string{"valid", "oversize", "redirect", "error"} {
		t.Run(scenario, func(t *testing.T) {
			u := &avatarUpload{}
			importer := NewAvatarImporter(u)
			importer.client.Transport = avatarTransport(func(r *http.Request) (*http.Response, error) {
				body := "image bytes"
				status := 200
				header := http.Header{}
				if scenario == "oversize" {
					body = strings.Repeat("x", int(application.MaxAvatarBytes)+1)
				}
				if scenario == "redirect" {
					status = 302
					header.Set("Location", "https://127.0.0.1/private")
				}
				if scenario == "error" {
					status = 503
				}
				return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})
			err := importer.ImportGoogleAvatar(context.Background(), "user", "https://lh3.googleusercontent.com/a/photo")
			if scenario == "valid" {
				if err != nil || u.calls != 1 || u.body != "image bytes" {
					t.Fatalf("image not imported: %v", err)
				}
			} else if err == nil || u.calls != 0 {
				t.Fatalf("unsafe response imported: %v", err)
			}
		})
	}
}

func TestGoogleAvatarRejectsNonPublicAddresses(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.64.0.1", "0.1.2.3", "::1", "fc00::1", "fe80::1"} {
		if publicAvatarIP(net.ParseIP(address)) {
			t.Errorf("non-public address accepted: %s", address)
		}
	}
	for _, address := range []string{"142.250.1.1", "2607:f8b0:4005:805::200e"} {
		if !publicAvatarIP(net.ParseIP(address)) {
			t.Errorf("public address rejected: %s", address)
		}
	}
}
