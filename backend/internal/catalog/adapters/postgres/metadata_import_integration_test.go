package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	platform "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

func TestMetadataImportCanonicalTagsCandidatesAndReview(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	base, err := platform.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "imports_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
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
	if _, err = db.Pool().Exec(ctx, `INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','admin@imports.test','test','admin',true),('reader','reader@imports.test','test','reader',true);INSERT INTO genres(id,name_kk,name_ru,slug) VALUES('import-shared-a','Ортақ','Общее','import-shared-a'),('import-shared-b','Ортақ','Общее','import-shared-b')`); err != nil {
		t.Fatal(err)
	}
	repo := New(db.Pool())
	if err = repo.AuthorizeImport(ctx, "reader", ""); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal(err)
	}
	tags := []application.ExternalTag{{ID: "stable-action", Name: "Action"}, {ID: "stable-action", Name: "Экшен"}, {ID: "unknown", Name: "Unknown external tag"}, {ID: "ambiguous", Name: "Общее"}}
	genres, categories, candidates, err := repo.MatchImportTags(ctx, "shikimori", tags)
	if err != nil {
		t.Fatal(err)
	}
	if len(genres) != 1 || len(categories) != 0 || len(candidates) != 2 {
		t.Fatalf("genres=%v categories=%v candidates=%+v", genres, categories, candidates)
	}
	var genre string
	if err = db.Pool().QueryRow(ctx, `SELECT id FROM genres WHERE slug='action'`).Scan(&genre); err != nil {
		t.Fatal(err)
	}
	if genres[0] != genre {
		t.Fatal(genres)
	}
	service := application.NewMetadataImports(repo, nil)
	if err = service.Resolve(ctx, "admin", "shikimori", "unknown", "genres", genre); err != nil {
		t.Fatal(err)
	}
	genres, _, candidates, err = repo.MatchImportTags(ctx, "shikimori", []application.ExternalTag{{ID: "unknown", Name: "Changed upstream name"}, {ID: "stable-action", Name: "Action"}})
	if err != nil || len(genres) != 1 || len(candidates) != 0 {
		t.Fatalf("stable mapping lost: %v %+v %v", genres, candidates, err)
	}
	if err = repo.RecordImport(ctx, "admin", "", application.ImportTitle{Source: "shikimori", ID: "13", URL: "https://shikimori.one/mangas/13", Name: "One Piece"}); err != nil {
		t.Fatal(err)
	}
	queue, err := repo.ImportCandidates(ctx, "admin")
	if err != nil || len(queue) != 1 || queue[0].Status != "ambiguous" {
		t.Fatalf("queue %+v %v", queue, err)
	}
	var receipts int
	if err = db.Pool().QueryRow(ctx, `SELECT count(*) FROM metadata_import_receipts WHERE source='shikimori' AND external_id='13'`).Scan(&receipts); err != nil || receipts != 1 {
		t.Fatal(receipts, err)
	}
}
