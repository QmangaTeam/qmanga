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

func TestAccountManagement(tester *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		tester.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, failure := databaseAdapter.New(ctx, connection)
	if failure != nil {
		tester.Fatal(failure)
	}
	defer base.Close()
	schema := "account_management_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, failure = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); failure != nil {
		tester.Fatal(failure)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(connection)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	database, failure := databaseAdapter.New(ctx, parsed.String())
	if failure != nil {
		tester.Fatal(failure)
	}
	defer database.Close()
	if failure = database.Migrate(ctx); failure != nil {
		tester.Fatal(failure)
	}
	if _, failure = database.Pool().Exec(ctx, "INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','admin@test.local','hash','admin',true),('mod','mod@test.local','hash','moderator',true),('reader','reader@test.local','hash','reader',true)"); failure != nil {
		tester.Fatal(failure)
	}
	repository := New(database.Pool())
	if failure = repository.ManageAccount(ctx, "reader", "mod", application.AccountChange{Role: "admin", Reason: "not permitted", Version: 0}); !errors.Is(failure, domain.ErrForbidden) {
		tester.Fatal("reader elevated privileges", failure)
	}
	if failure = repository.ManageAccount(ctx, "admin", "admin", application.AccountChange{Role: "reader", Reason: "demote self", Version: 0}); !errors.Is(failure, domain.ErrForbidden) {
		tester.Fatal("last admin lost", failure)
	}
	if failure = repository.ManageAccount(ctx, "mod", "reader", application.AccountChange{Role: "admin", Reason: "not permitted", Version: 0}); !errors.Is(failure, domain.ErrForbidden) {
		tester.Fatal("moderator granted admin", failure)
	}
	if failure = repository.SaveSession(ctx, application.Session{Hash: "session", UserID: "reader", CSRFHash: "csrf", ExpiresAt: time.Now().Add(time.Hour)}); failure != nil {
		tester.Fatal(failure)
	}
	if failure = repository.ManageAccount(ctx, "mod", "reader", application.AccountChange{Role: "reader", Blocked: true, Reason: "Repeated spam messages", Version: 0}); failure != nil {
		tester.Fatal(failure)
	}
	if _, _, failure = repository.SessionAccount(ctx, "session", time.Now()); !errors.Is(failure, domain.ErrCredentials) {
		tester.Fatal("blocked session remains valid", failure)
	}
	if _, failure = repository.AccountByEmail(ctx, "reader@test.local"); !errors.Is(failure, domain.ErrNotFound) {
		tester.Fatal("blocked login allowed", failure)
	}
	if failure = repository.ManageAccount(ctx, "admin", "reader", application.AccountChange{Role: "reader", Reason: "stale unblock", Version: 0}); !errors.Is(failure, domain.ErrConflict) {
		tester.Fatal("stale change overwrote block", failure)
	}
	history, failure := repository.Audit(ctx, "admin", "")
	if failure != nil || len(history) != 1 {
		tester.Fatal("missing management audit", history, failure)
	}
	if _, failure = repository.Audit(ctx, "reader", ""); !errors.Is(failure, domain.ErrForbidden) {
		tester.Fatal("audit leaked", failure)
	}
}
