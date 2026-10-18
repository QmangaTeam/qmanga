package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/url"
	"os"
	"qmanga/internal/identity/application"
	platform "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

type guardedAvatarStorage struct {
	pool          *pgxpool.Pool
	deleted       int
	guardObserved bool
}

func (s *guardedAvatarStorage) DeleteDerivative(ctx context.Context, _ string) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	var acquired bool
	if err = conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(482161,1)`).Scan(&acquired); err != nil {
		return err
	}
	s.guardObserved = !acquired
	if acquired {
		_, _ = conn.Exec(ctx, `SELECT pg_advisory_unlock(482161,1)`)
	}
	s.deleted++
	return nil
}
func (*guardedAvatarStorage) PutDerivative(context.Context, string, io.Reader, int64, string) error {
	return nil
}
func (*guardedAvatarStorage) OpenDerivative(context.Context, string, int64, int64) (io.ReadCloser, error) {
	return nil, nil
}

func TestBackupGuardSuppressesAvatarDeletionAndCleanupResumes(t *testing.T) {
	rawURL := os.Getenv("TEST_DATABASE_URL")
	if rawURL == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := platform.New(ctx, rawURL)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "avatar_backup_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(rawURL)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := platform.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	repo := New(db.Pool())
	key := "avatars/obsolete"
	if err = repo.ReserveAvatarObject(ctx, key); err != nil {
		t.Fatal(err)
	}
	if err = repo.AbandonAvatarObject(ctx, key); err != nil {
		t.Fatal(err)
	}
	backup, err := db.Pool().Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer backup.Release()
	if _, err = backup.Exec(ctx, `SELECT pg_advisory_lock(482161,1)`); err != nil {
		t.Fatal(err)
	}
	defer backup.Exec(context.Background(), `SELECT pg_advisory_unlock_all()`)
	storage := &guardedAvatarStorage{pool: db.Pool()}
	service := application.NewAvatars(repo, storage, nil, nil, nil)
	if err = service.DrainAvatarGarbage(ctx); err != nil || storage.deleted != 0 {
		t.Fatal("backup permitted physical deletion", err, storage.deleted)
	}
	queued, err := repo.AvatarGarbage(ctx)
	if err != nil || len(queued) != 1 {
		t.Fatal("backup lost deletion receipt", queued, err)
	}
	if _, err = backup.Exec(ctx, `SELECT pg_advisory_unlock(482161,1)`); err != nil {
		t.Fatal(err)
	}
	if err = service.DrainAvatarGarbage(ctx); err != nil || storage.deleted != 1 || !storage.guardObserved {
		t.Fatal("unguarded/missing resumed cleanup", err, storage.deleted, storage.guardObserved)
	}
	queued, err = repo.AvatarGarbage(ctx)
	if err != nil || len(queued) != 0 {
		t.Fatal("completed cleanup retained receipt", queued, err)
	}
	var available bool
	if err = backup.QueryRow(ctx, `SELECT pg_try_advisory_lock(482161,1)`).Scan(&available); err != nil || !available {
		t.Fatal("collector leaked guard", err, available)
	}
}
