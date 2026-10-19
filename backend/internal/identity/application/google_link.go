package application

import (
	"context"
	"qmanga/internal/identity/domain"
	"time"
)

type GoogleLinkRepository interface {
	LinkGoogleIdentity(context.Context, string, string, GoogleIdentity, time.Time) error
	GoogleLinked(context.Context, string) (bool, error)
}

func (s *GoogleService) GoogleLinked(ctx context.Context, actor string) (bool, error) {
	repository, ok := s.repository.(GoogleLinkRepository)
	if !ok || actor == "" {
		return false, domain.ErrForbidden
	}
	return repository.GoogleLinked(ctx, actor)
}
func (s *GoogleService) StartLink(ctx context.Context, actor, sessionHash string) (string, string, error) {
	if actor == "" || sessionHash == "" || s.identity == nil {
		return "", "", domain.ErrForbidden
	}
	session, account, err := s.identity.repository.SessionAccount(ctx, sessionHash, s.now())
	if err != nil || account.ID != actor || session.UserID != actor || !account.Verified {
		return "", "", domain.ErrForbidden
	}
	return s.start(ctx, "/settings", "", actor, sessionHash)
}
func (s *GoogleService) FinishLink(ctx context.Context, state, code, sessionToken string) error {
	if len(state) != 43 || s.provider == nil {
		return domain.ErrCredentials
	}
	flow, err := s.repository.ConsumeGoogleFlow(ctx, s.credentials.Digest(state), s.now())
	if err != nil {
		return err
	}
	if flow.LinkUserID == "" || flow.LinkSessionHash == "" || s.identity == nil {
		return domain.ErrForbidden
	}
	session, account, err := s.identity.Authenticate(ctx, sessionToken)
	if err != nil || account.ID != flow.LinkUserID || session.Hash != flow.LinkSessionHash || session.UserID != flow.LinkUserID || !account.Verified {
		return domain.ErrForbidden
	}
	if code == "" {
		return domain.ErrCredentials
	}
	identity, err := s.provider.Exchange(ctx, code, flow.Nonce, flow.Verifier)
	if err != nil {
		return domain.ErrCredentials
	}
	email, err := domain.NormalizeEmail(identity.Email)
	if err != nil || !identity.Verified || identity.Subject == "" || identity.Issuer != "https://accounts.google.com" {
		return domain.ErrCredentials
	}
	identity.Email = email
	repository, ok := s.repository.(GoogleLinkRepository)
	if !ok {
		return domain.ErrForbidden
	}
	// Repository must recheck the exact live session and verified actor in the insert transaction.
	return repository.LinkGoogleIdentity(ctx, flow.LinkUserID, flow.LinkSessionHash, identity, s.now())
}
