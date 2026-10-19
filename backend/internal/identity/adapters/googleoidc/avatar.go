package googleoidc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"qmanga/internal/identity/application"
	"time"
)

type avatarUploader interface {
	Upload(context.Context, string, string, string, io.Reader) (application.Avatar, error)
}
type AvatarImporter struct {
	avatars avatarUploader
	client  *http.Client
}

func NewAvatarImporter(avatars avatarUploader) *AvatarImporter {
	transport := &http.Transport{TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 5 * time.Second, DialContext: googleAvatarDial}
	return &AvatarImporter{avatars: avatars, client: &http.Client{Transport: transport, Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

// Resolve once and dial the checked address, so DNS cannot redirect a second lookup into private infrastructure.
func googleAvatarDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || port != "443" {
		return nil, fmt.Errorf("invalid avatar address")
	}
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, address := range addresses {
		if !publicAvatarIP(address.IP) {
			return nil, fmt.Errorf("non-public avatar address")
		}
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	for _, address := range addresses {
		conn, failure := dialer.DialContext(ctx, network, net.JoinHostPort(address.IP.String(), port))
		if failure == nil {
			return conn, nil
		}
		err = failure
	}
	if err == nil {
		err = fmt.Errorf("no avatar address")
	}
	return nil, err
}
func publicAvatarIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	for _, prefix := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4", "2001:db8::/32"} {
		if netip.MustParsePrefix(prefix).Contains(address) {
			return false
		}
	}
	return true
}
func (i *AvatarImporter) ImportGoogleAvatar(ctx context.Context, owner, source string) error {
	parsed, err := url.Parse(source)
	if err != nil || len(source) > 2048 || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" || parsed.Opaque != "" {
		return fmt.Errorf("invalid Google avatar URL")
	}
	switch parsed.Host {
	case "lh3.googleusercontent.com", "lh4.googleusercontent.com", "lh5.googleusercontent.com", "lh6.googleusercontent.com":
	default:
		return fmt.Errorf("untrusted Google avatar host")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return err
	}
	response, err := i.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.ContentLength > application.MaxAvatarBytes {
		return fmt.Errorf("unavailable Google avatar")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, application.MaxAvatarBytes+1))
	if err != nil {
		return err
	}
	if len(raw) == 0 || int64(len(raw)) > application.MaxAvatarBytes {
		return fmt.Errorf("Google avatar size exceeded")
	}
	_, err = i.avatars.Upload(ctx, owner, "user", owner, bytes.NewReader(raw))
	return err
}
