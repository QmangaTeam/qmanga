package postgres

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	identityCrypto "qmanga/internal/identity/adapters/crypto"
	"qmanga/internal/platform/postgres"
	"qmanga/internal/teams/application"
	"qmanga/internal/teams/domain"
)

type capturedMail struct{ body string }

func (mailer *capturedMail) Send(ctx context.Context, to, subject, body string) error {
	mailer.body = body
	return nil
}

func TestTeamIsolationAndRevocation(testingTarget *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		testingTarget.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, failure := postgres.New(ctx, databaseURL)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer base.Close()
	schema := "teams_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, failure = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); failure != nil {
		testingTarget.Fatal(failure)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
	parameters := parsed.Query()
	parameters.Set("search_path", schema)
	parsed.RawQuery = parameters.Encode()
	database, failure := postgres.New(ctx, parsed.String())
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer database.Close()
	if failure = database.Migrate(ctx); failure != nil {
		testingTarget.Fatal(failure)
	}
	for _, account := range []struct{ id, email, role string }{
		{"owner", "owner@example.test", "reader"}, {"other", "other@example.test", "reader"}, {"member", "member@example.test", "reader"}, {"admin", "admin@example.test", "admin"},
	} {
		if _, failure = database.Pool().Exec(ctx, "INSERT INTO users(id,email,password_hash,role,is_verified)VALUES($1,$2,'test-only',$3,true)", account.id, account.email, account.role); failure != nil {
			testingTarget.Fatal(failure)
		}
	}
	mailer := &capturedMail{}
	service := application.New(New(database.Pool()), identityCrypto.New(), mailer, "http://localhost:8080")
	team, failure := service.Create(ctx, "owner", "dala", "Дала тобы", "Қазақша аудармалар")
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if !errors.Is(service.Check(ctx, "owner", team.ID, domain.Upload), domain.ErrForbidden) {
		testingTarget.Fatal("unapproved upload allowed")
	}
	if !errors.Is(service.Approve(ctx, "owner", team.ID), domain.ErrForbidden) {
		testingTarget.Fatal("owner approved own team")
	}
	if failure = service.Approve(ctx, "admin", team.ID); failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure = service.Check(ctx, "owner", team.ID, domain.Submit); failure != nil {
		testingTarget.Fatal(failure)
	}
	if !errors.Is(service.Check(ctx, "other", team.ID, domain.ReadDraft), domain.ErrForbidden) {
		testingTarget.Fatal("foreign draft exposed")
	}
	if !errors.Is(service.Invite(ctx, "other", team.ID, "member@example.test", domain.Editor), domain.ErrForbidden) {
		testingTarget.Fatal("foreign invite accepted")
	}
	if failure = service.Invite(ctx, "owner", team.ID, "member@example.test", domain.Editor); failure != nil {
		testingTarget.Fatal(failure)
	}
	parts := strings.Split(mailer.body, "token=")
	token := parts[1][:43]
	if _, failure = service.Accept(ctx, "other", token); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("invitation accepted by wrong email")
	}
	if _, failure = service.Accept(ctx, "member", token); failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = service.Accept(ctx, "member", token); !errors.Is(failure, domain.ErrExpired) {
		testingTarget.Fatal("invitation reused")
	}
	if failure = service.Check(ctx, "member", team.ID, domain.Submit); failure != nil {
		testingTarget.Fatal(failure)
	}
	if !errors.Is(service.Remove(ctx, "owner", team.ID, "owner"), domain.ErrConflict) {
		testingTarget.Fatal("last owner removed")
	}
	if failure = service.Remove(ctx, "owner", team.ID, "member"); failure != nil {
		testingTarget.Fatal(failure)
	}
	if !errors.Is(service.Check(ctx, "member", team.ID, domain.ReadDraft), domain.ErrForbidden) {
		testingTarget.Fatal("removed member retains access")
	}
}
