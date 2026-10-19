package googleoidc

import (
	"context"
	"crypto/subtle"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"net/http"
	"qmanga/internal/identity/application"
	"time"
)

type Provider struct {
	config   oauth2.Config
	verifier *oidc.IDTokenVerifier
	client   *http.Client
}

func New(clientID, secret, redirect string) *Provider {
	client := &http.Client{Timeout: 12 * time.Second}
	keys := oidc.NewRemoteKeySet(oidc.ClientContext(context.Background(), client), "https://www.googleapis.com/oauth2/v3/certs")
	return &Provider{config: oauth2.Config{ClientID: clientID, ClientSecret: secret, RedirectURL: redirect, Scopes: []string{"openid", "email", "profile"}, Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/v2/auth", TokenURL: "https://oauth2.googleapis.com/token"}}, verifier: oidc.NewVerifier("https://accounts.google.com", keys, &oidc.Config{ClientID: clientID, SupportedSigningAlgs: []string{"RS256"}}), client: client}
}
func (p *Provider) AuthorizationURL(state, nonce, verifier string) string {
	return p.config.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier))
}
func (p *Provider) Exchange(ctx context.Context, code, nonce, verifier string) (application.GoogleIdentity, error) {
	ctx = oidc.ClientContext(ctx, p.client)
	token, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return application.GoogleIdentity{}, err
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok {
		return application.GoogleIdentity{}, fmt.Errorf("missing id token")
	}
	id, err := p.verifier.Verify(ctx, raw)
	if err != nil {
		return application.GoogleIdentity{}, err
	}
	if len(nonce) < 32 || subtle.ConstantTimeCompare([]byte(id.Nonce), []byte(nonce)) != 1 {
		return application.GoogleIdentity{}, fmt.Errorf("invalid nonce")
	}
	var claims struct {
		Email           string `json:"email"`
		Name            string `json:"name"`
		Picture         string `json:"picture"`
		EmailVerified   bool   `json:"email_verified"`
		AuthorizedParty string `json:"azp"`
	}
	if err = id.Claims(&claims); err != nil {
		return application.GoogleIdentity{}, err
	}
	if claims.AuthorizedParty != "" && claims.AuthorizedParty != p.config.ClientID {
		return application.GoogleIdentity{}, fmt.Errorf("invalid authorized party")
	}
	return application.GoogleIdentity{Issuer: "https://accounts.google.com", Subject: id.Subject, Email: claims.Email, Verified: claims.EmailVerified, Name: claims.Name, Picture: claims.Picture}, nil
}
