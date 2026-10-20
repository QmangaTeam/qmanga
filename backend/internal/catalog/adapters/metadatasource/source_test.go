package metadatasource

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"qmanga/internal/catalog/application"
	"testing"
	"time"
)

func TestShikimoriFixtureCacheAndIdentity(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/mangas" || r.URL.Query().Get("search") != "One Piece" {
			t.Errorf("unexpected request %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":13,"name":"One Piece","russian":"Ван-Пис","kind":"manga","status":"ongoing","aired_on":"1997-07-22","image":{"original":"/cover.jpg"}}]`))
	}))
	defer server.Close()
	source := New("shikimori")
	source.endpoint = server.URL
	items, err := source.Search(context.Background(), "One Piece")
	if err != nil || len(items) != 1 || items[0].ID != "13" || items[0].Year != 1997 || items[0].Kind != "manga" {
		t.Fatalf("%+v %v", items, err)
	}
	if _, err = source.Search(context.Background(), "One Piece"); err != nil || calls != 1 {
		t.Fatal("missing cache", err, calls)
	}
	if _, err = source.Search(context.Background(), "other"); !errors.Is(err, application.ErrSourceLimited) {
		t.Fatal("missing source limit", err)
	}
}
func TestAniListFixturePartialAndSanitization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error(r.Method)
		}
		w.Write([]byte(`{"data":{"Media":{"id":30013,"title":{"romaji":"One Piece","english":"ONE PIECE","native":"ワンピース"},"description":"<b>Story</b>","format":"MANGA","status":"RELEASING","countryOfOrigin":"JP","startDate":{"year":1997},"coverImage":{"large":"https://evil.example/track"},"genres":["Action"],"tags":[{"id":1,"name":"Pirates"}],"staff":{"edges":[{"role":"Story & Art","node":{"name":{"full":"Eiichiro Oda"}}}]}}}}`))
	}))
	defer server.Close()
	source := New("anilist")
	source.endpoint = server.URL
	item, err := source.Detail(context.Background(), "30013")
	if err != nil || item.CoverURL != "" || item.Description != "Story" || len(item.Authors) != 1 || len(item.Tags) != 2 || item.Tags[0].ID != "genre:Action" {
		t.Fatalf("%+v %v", item, err)
	}
}
func TestSourceFailureAndRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Retry-After", "10"); w.WriteHeader(429) }))
	defer server.Close()
	source := New("anilist")
	source.endpoint = server.URL
	_, err := source.Search(context.Background(), "test")
	if !errors.Is(err, application.ErrSourceLimited) || time.Until(source.next) < 9*time.Second {
		t.Fatal(err, source.next)
	}
	other := New("shikimori")
	if !other.next.IsZero() {
		t.Fatal("source budgets mixed")
	}
}
func TestUnknownKindAndUnsafeArtworkStayUnknown(t *testing.T) {
	s := New("shikimori")
	result := s.shikiMap(shikiTitle{ID: 1, Name: "Example", Kind: "novel", Status: "announced"})
	if result.Kind != "" || result.OriginalStatus != "" || result.Year != 0 {
		t.Fatal(result)
	}
	for _, value := range []string{"http://s4.anilist.co/a", "https://user@s4.anilist.co/a", "https://s4.anilist.co.evil/a", "https://s4.anilist.co:999/a"} {
		if safeImage(value, "s4.anilist.co") != "" {
			t.Fatal(value)
		}
	}
}
func TestAniListDetailPreservesRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(429) }))
	defer server.Close()
	source := New("anilist")
	source.endpoint = server.URL
	if _, err := source.Detail(context.Background(), "13"); !errors.Is(err, application.ErrSourceLimited) {
		t.Fatal(err)
	}
}
