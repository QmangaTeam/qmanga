package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	databaseAdapter "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

func TestGooglePersistenceAndAvatarVisibility(t *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := databaseAdapter.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "google_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(connection)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := databaseAdapter.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	repo := New(db.Pool())
	flow := application.GoogleFlow{Hash: "state", Nonce: "nonce", Verifier: "verifier", ReturnPath: "/catalog", ExpiresAt: time.Now().Add(time.Minute)}
	if err = repo.SaveGoogleFlow(ctx, flow); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ConsumeGoogleFlow(ctx, "state", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ConsumeGoogleFlow(ctx, "state", time.Now()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("state replay accepted", err)
	}
	identity := application.GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "subject", Email: "google@test.local", Verified: true}
	candidate := domain.Account{ID: "google-user", Email: identity.Email, PasswordHash: "opaque-random-hash", DisplayName: "Reader", Role: "reader", Verified: true}
	account, err := repo.GoogleAccount(ctx, identity, candidate)
	if err != nil || account.ID != candidate.ID {
		t.Fatal("new Google account", err)
	}
	if err = repo.UpdateProfile(ctx, account.ID, "Edited nickname"); err != nil {
		t.Fatal(err)
	}
	candidate.DisplayName = "Changed Google name"
	candidate.ID = "other"
	candidate.Email = "different@test.local"
	again, err := repo.GoogleAccount(ctx, identity, candidate)
	if err != nil || again.ID != account.ID || again.DisplayName != "Edited nickname" || again.Email != account.Email {
		t.Fatal("stable subject lost", err)
	}
	// One user keeps exactly one Google identity, so a second subject with the same email conflicts.
	identity.Subject = "attacker"
	if _, err = repo.GoogleAccount(ctx, identity, candidate); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("second Google identity accepted", err)
	}

	// Registration import receipts are created atomically only for new accounts.
	now := time.Now()
	fresh := domain.Account{ID: "registration-new", Email: "registration-new@example.test", PasswordHash: "hash", Role: "reader", Verified: true, DisplayName: "New reader"}
	receipt := strings.Repeat("a", 64)
	if err = repo.CreateAccountForImport(ctx, fresh, receipt, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		actor, hash string
		want        bool
	}{{fresh.ID, receipt, true}, {account.ID, receipt, false}, {fresh.ID, strings.Repeat("b", 64), false}} {
		allowed, failure := repo.RegistrationImportAllowed(ctx, check.actor, check.hash, now)
		if failure != nil || allowed != check.want {
			t.Fatal(check.actor, allowed, failure)
		}
	}
	duplicate := fresh
	duplicate.ID = "registration-duplicate"
	if err = repo.CreateAccountForImport(ctx, duplicate, strings.Repeat("b", 64), now.Add(time.Hour)); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("duplicate registration created receipt", err)
	}
	if err = repo.CompleteRegistrationImport(ctx, account.ID, receipt, now); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("foreign completion accepted", err)
	}
	if err = repo.CompleteRegistrationImport(ctx, fresh.ID, receipt, now); err != nil {
		t.Fatal(err)
	}
	if err = repo.CompleteRegistrationImport(ctx, fresh.ID, receipt, now); err != nil {
		t.Fatal("completion replay must be harmless", err)
	}
	if allowed, failure := repo.RegistrationImportAllowed(ctx, fresh.ID, receipt, now); failure != nil || allowed {
		t.Fatal("receipt replay", allowed, failure)
	}
	googleIdentity := application.GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "registration-new-google", Email: "registration-google@example.test", Verified: true}
	googleCandidate := domain.Account{ID: "registration-google", Email: googleIdentity.Email, PasswordHash: "hash", Role: "reader", Verified: true, DisplayName: "Google reader"}
	if _, err = repo.GoogleAccountForImport(ctx, googleIdentity, googleCandidate, receipt, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if allowed, failure := repo.RegistrationImportAllowed(ctx, googleCandidate.ID, receipt, now); failure != nil || !allowed {
		t.Fatal("new Google import missing", allowed, failure)
	}
	if _, err = repo.GoogleAccountForImport(ctx, googleIdentity, googleCandidate, strings.Repeat("b", 64), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if allowed, failure := repo.RegistrationImportAllowed(ctx, googleCandidate.ID, strings.Repeat("b", 64), now); failure != nil || allowed {
		t.Fatal("ordinary Google login granted import", allowed, failure)
	}
	if allowed, failure := repo.RegistrationImportAllowed(ctx, googleCandidate.ID, receipt, now.Add(2*time.Hour)); failure != nil || allowed {
		t.Fatal("expired receipt accepted", allowed, failure)
	}

	// Explicit linking preserves an ordinary account and its origin avatar; email alone still refuses access.
	owner := domain.Account{ID: "link-owner", Email: "owner@example.test", PasswordHash: "hash", Role: "admin", Verified: true, DisplayName: "My edited profile"}
	foreign := domain.Account{ID: "link-foreign", Email: "foreign@example.test", PasswordHash: "hash", Role: "reader", Verified: true, DisplayName: "Foreign reader"}
	for _, actor := range []domain.Account{owner, foreign} {
		if err = repo.CreateAccount(ctx, actor); err != nil {
			t.Fatal(err)
		}
		if err = repo.SaveSession(ctx, application.Session{Hash: actor.ID + "-session", UserID: actor.ID, CSRFHash: "csrf", ExpiresAt: now.Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	linkedIdentity := application.GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "explicit-link", Email: owner.Email, Verified: true, Name: "Google changed name"}
	// A verified Google email links a password-registered account and signs it in (owner decision 19 October 2026).
	if linked, err := repo.GoogleLinked(ctx, owner.ID); err != nil || linked {
		t.Fatal("unlinked status", err)
	}
	if actualGoogle, err := repo.GoogleAccount(ctx, linkedIdentity, domain.Account{ID: "must-not-create", Email: owner.Email, DisplayName: "Google name"}); err != nil || actualGoogle != owner {
		t.Fatalf("same-email identity did not auto-link: %+v %v", actualGoogle, err)
	}
	customAvatar := application.Avatar{Kind: "user", OwnerID: owner.ID, Key: "avatars/custom-link-owner", Bytes: 100, UpdatedAt: now}
	if err = repo.SaveAvatar(ctx, customAvatar); err != nil {
		t.Fatal(err)
	}
	if linked, err := repo.GoogleLinked(ctx, owner.ID); err != nil || !linked {
		t.Fatal("auto-link status", err)
	}
	if actual, err := repo.GoogleAccount(ctx, linkedIdentity, domain.Account{ID: "new-candidate", Email: "changed@example.test", DisplayName: "New name"}); err != nil || actual != owner {
		t.Fatalf("linked login changed account: %+v %v", actual, err)
	}
	if actual, err := repo.Avatar(ctx, "user", owner.ID); err != nil || actual.Key != customAvatar.Key {
		t.Fatal("link replaced avatar", err)
	}
	if err = repo.LinkGoogleIdentity(ctx, foreign.ID, foreign.ID+"-session", linkedIdentity, now); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("foreign identity transferred", err)
	}
	if err = repo.LinkGoogleIdentity(ctx, owner.ID, owner.ID+"-session", application.GoogleIdentity{Issuer: linkedIdentity.Issuer, Subject: "replacement-google", Email: owner.Email, Verified: true}, now); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("existing Google replaced", err)
	}
	for _, condition := range []string{"wrong-session", "expired", "revoked", "unverified", "blocked"} {
		t.Run(condition, func(t *testing.T) {
			sessionHash := foreign.ID + "-session"
			if err = repo.SaveSession(ctx, application.Session{Hash: "temporary-link-session", UserID: foreign.ID, CSRFHash: "csrf", ExpiresAt: now.Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			sessionHash = "temporary-link-session"
			switch condition {
			case "wrong-session":
				sessionHash = owner.ID + "-session"
			case "expired":
				_, err = db.Pool().Exec(ctx, "UPDATE sessions SET expires_at=$1 WHERE id=$2", now.Add(-time.Second), sessionHash)
			case "revoked":
				err = repo.DeleteSession(ctx, sessionHash)
			case "unverified":
				_, err = db.Pool().Exec(ctx, "UPDATE users SET is_verified=false WHERE id=$1", foreign.ID)
			case "blocked":
				_, err = db.Pool().Exec(ctx, "UPDATE users SET is_blocked=true WHERE id=$1", foreign.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			if failure := repo.LinkGoogleIdentity(ctx, foreign.ID, sessionHash, application.GoogleIdentity{Issuer: linkedIdentity.Issuer, Subject: "invalid-link-" + condition, Email: foreign.Email, Verified: true}, now); !errors.Is(failure, domain.ErrForbidden) {
				t.Fatal("invalid session accepted", failure)
			}
			if _, err = db.Pool().Exec(ctx, "UPDATE users SET is_verified=true,is_blocked=false WHERE id=$1", foreign.ID); err != nil {
				t.Fatal(err)
			}
			if err = repo.DeleteSession(ctx, "temporary-link-session"); err != nil {
				t.Fatal(err)
			}
		})
	}
	avatar := application.Avatar{Kind: "user", OwnerID: account.ID, Key: "avatars/immutable", Bytes: 100, UpdatedAt: time.Now()}
	if err = repo.SaveAvatar(ctx, avatar); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Avatar(ctx, "user", account.ID); err != nil {
		t.Fatal(err)
	}

	// Replacement queues only the obsolete object, and even an accidental current-key queue is excluded.
	replacement := avatar
	replacement.Key = "avatars/new"
	if err = repo.ReserveAvatarObject(ctx, replacement.Key); err != nil {
		t.Fatal(err)
	}
	if err = repo.SaveAvatar(ctx, replacement); err != nil {
		t.Fatal(err)
	}
	if err = repo.ReserveAvatarObject(ctx, replacement.Key); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Pool().Exec(ctx, "UPDATE avatar_garbage SET ready_at=now()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	garbage, err := repo.AvatarGarbage(ctx)
	if err != nil || len(garbage) != 1 || garbage[0] != avatar.Key {
		t.Fatalf("current avatar included in cleanup: %v %v", garbage, err)
	}
	if _, err = db.Pool().Exec(ctx, "UPDATE users SET is_blocked=true WHERE id=$1", account.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Avatar(ctx, "user", account.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("blocked avatar remained public", err)
	}
	identity.Subject = "subject"
	if _, err = repo.GoogleAccount(ctx, identity, candidate); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("blocked Google identity accepted", err)
	}
}

func TestGoogleEmailMatchLinksRegisteredAccount(t *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := databaseAdapter.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "google_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsedURL, _ := url.Parse(connection)
	query := parsedURL.Query()
	query.Set("search_path", schema)
	parsedURL.RawQuery = query.Encode()
	db, err := databaseAdapter.New(ctx, parsedURL.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	repo := New(db.Pool())
	password := domain.Account{ID: "pwd-user", Email: "pwd@example.test", PasswordHash: "hash", Role: "reader", Verified: true, DisplayName: "Password reader"}
	if err := repo.CreateAccount(ctx, password); err != nil {
		t.Fatal(err)
	}
	identity := application.GoogleIdentity{Issuer: "https://accounts.google.com", Subject: "email-link-sub", Email: "pwd@example.test", Verified: true}
	linked, err := repo.GoogleAccount(ctx, identity, domain.Account{ID: "unused", Email: "pwd@example.test", PasswordHash: "opaque", DisplayName: "Google override", Role: "reader", Verified: true})
	if err != nil || linked.ID != password.ID {
		t.Fatalf("auto-link missing: %+v %v", linked, err)
	}
	if linked != password {
		t.Fatalf("linked account changed: %+v", linked)
	}
	if again, err := repo.GoogleAccount(ctx, identity, domain.Account{ID: "unused", Email: "pwd@example.test", DisplayName: "x"}); err != nil || again.ID != password.ID {
		t.Fatalf("repeat login after link: %+v %v", again, err)
	}
	var users int
	if err = db.Pool().QueryRow(ctx, "SELECT count(*) FROM users WHERE lower(email)=$1", password.Email).Scan(&users); err != nil || users != 1 {
		t.Fatalf("users for email: %d %v", users, err)
	}
	if _, err = repo.GoogleAccount(ctx, application.GoogleIdentity{Issuer: identity.Issuer, Subject: "another-sub", Email: password.Email, Verified: true}, domain.Account{ID: "unused", Email: password.Email, DisplayName: "y"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second subject accepted: %v", err)
	}
	if _, err = db.Pool().Exec(ctx, "UPDATE users SET is_blocked=true WHERE id=$1", password.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.GoogleAccount(ctx, identity, domain.Account{ID: "unused", Email: password.Email, DisplayName: "z"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("blocked email match accepted: %v", err)
	}
}

