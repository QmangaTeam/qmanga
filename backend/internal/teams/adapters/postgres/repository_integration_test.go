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

	testingTarget.Run("team and member lists expose later pages without overlap", func(testingTarget *testing.T) {
		for _, statement := range []string{
			"INSERT INTO teams(id,slug,name,description) SELECT 'paged-team-'||number,'paged-team-'||number,'Paged team '||number,'' FROM generate_series(1,101) number",
			"INSERT INTO team_memberships(id,team_id,user_id,role) SELECT 'paged-owner-'||number,'paged-team-'||number,'owner','owner' FROM generate_series(1,101) number",
			"INSERT INTO users(id,email,password_hash,is_verified) SELECT 'paged-user-'||number,'paged-user-'||number||'@example.test','test-only',true FROM generate_series(1,101) number",
			"INSERT INTO team_memberships(id,team_id,user_id,role) SELECT 'paged-member-'||number,'paged-team-1','paged-user-'||number,'member' FROM generate_series(1,101) number",
		} {
			if _, failure := database.Pool().Exec(ctx, statement); failure != nil {
				testingTarget.Fatal(failure)
			}
		}
		first, failure := service.List(ctx, "owner", 0)
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		second, failure := service.List(ctx, "owner", 100)
		if failure != nil || len(first) != 100 || len(second) != 2 {
			testingTarget.Fatalf("team pages: %d/%d, %v", len(first), len(second), failure)
		}
		seen := map[string]bool{}
		for _, membership := range first {
			seen[membership.Team.ID] = true
		}
		for _, membership := range second {
			if seen[membership.Team.ID] {
				testingTarget.Fatal("team page overlaps")
			}
		}
		members, failure := service.Members(ctx, "owner", "paged-team-1", 0)
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		more, failure := service.Members(ctx, "owner", "paged-team-1", 100)
		if failure != nil || len(members) != 100 || len(more) != 2 {
			testingTarget.Fatalf("member pages: %d/%d, %v", len(members), len(more), failure)
		}
		seen = map[string]bool{}
		for _, member := range members {
			seen[member.UserID] = true
		}
		for _, member := range more {
			if seen[member.UserID] {
				testingTarget.Fatal("member page overlaps")
			}
		}
		if _, failure := service.Members(ctx, "other", "paged-team-1", 100); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("pagination bypasses membership authorization")
		}
	})

	testingTarget.Run("team rejection is visible and owner can correct and resubmit", func(testingTarget *testing.T) {
		requested, failure := service.Create(ctx, "owner", "review-team", "Review team", "Needs review")
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		if failure = service.Reject(ctx, "other", requested.ID, "Недостаточно сведений"); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("reader rejected team", failure)
		}
		if failure = service.Reject(ctx, "admin", requested.ID, "Недостаточно сведений"); failure != nil {
			testingTarget.Fatal(failure)
		}
		if failure = service.Approve(ctx, "admin", requested.ID); !errors.Is(failure, domain.ErrConflict) {
			testingTarget.Fatal("stale approval overrides rejected application", failure)
		}
		listed, failure := service.List(ctx, "owner")
		later, listFailure := service.List(ctx, "owner", 100)
		if listFailure != nil {
			testingTarget.Fatal(listFailure)
		}
		listed = append(listed, later...)
		found := false
		for _, item := range listed {
			if item.Team.ID == requested.ID {
				found = item.Team.ReviewStatus == "rejected" && item.Team.ReviewNote == "Недостаточно сведений"
			}
		}
		if failure != nil || !found {
			testingTarget.Fatal("owner cannot see review outcome", failure)
		}
		if failure = service.Update(ctx, "other", requested.ID, requested.Slug, requested.Name, "Fixed"); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("foreign team edited", failure)
		}
		if failure = service.Update(ctx, "owner", requested.ID, requested.Slug, requested.Name, "Подробное описание команды"); failure != nil {
			testingTarget.Fatal(failure)
		}
		pending, failure := service.Pending(ctx, "admin")
		laterPending, listFailure := service.Pending(ctx, "admin", 100)
		if listFailure != nil {
			testingTarget.Fatal(listFailure)
		}
		pending = append(pending, laterPending...)
		found = false
		for _, item := range pending {
			if item.ID == requested.ID {
				found = item.ReviewStatus == "pending" && item.ReviewNote == ""
			}
		}
		if failure != nil || !found {
			testingTarget.Fatal("resubmitted team absent from queue", failure)
		}
		if failure = service.Approve(ctx, "admin", requested.ID); failure != nil {
			testingTarget.Fatal(failure)
		}
	})
}
