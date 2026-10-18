package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"qmanga/internal/catalog/application"
)

type catalogueRepository struct {
	application.Repository
	filter application.Filter
}

func (repository *catalogueRepository) List(ctx context.Context, filter application.Filter) (application.Page, error) {
	repository.filter = filter
	return application.Page{Items: []application.Summary{{ID: "title", Name: "Дала үні", Slug: "dala", UpdatedAt: time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)}}, Next: &application.Cursor{ID: "title", Name: "Дала үні", UpdatedAt: time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)}}, nil
}

func TestCatalogueHTTPValidationAndCursor(testingTarget *testing.T) {
	repository := &catalogueRepository{}
	router := NewRouter(nil)
	NewCatalogHandler(application.New(repository, func() string { return "id" })).Register(router.Group("/api/v1"), NewIdentityHandler(nil, "http://localhost:8080", false))
	request := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/v1/"+path, nil))
		return recorder
	}
	for _, path := range []string{"titles?cursor=broken", "titles?limit=wat", "titles?original_status=private", "titles?limit=100000", "titles?age_rating=14", "titles?age_rating=-1", "titles?age_rating=nope"} {
		if response := request(path); response.Code != 400 && response.Code != 422 {
			testingTarget.Fatalf("invalid input accepted: %s %d", path, response.Code)
		}
	}
	response := request("titles?sort=name&limit=2")
	if response.Code != 200 || strings.Contains(response.Body.String(), "CoverID") {
		testingTarget.Fatal("invalid public DTO", response.Body.String())
	}
	var body struct {
		Next string `json:"next_cursor"`
	}
	if failure := json.Unmarshal(response.Body.Bytes(), &body); failure != nil || body.Next == "" {
		testingTarget.Fatal("missing cursor", failure)
	}
	if request("titles?sort=name&limit=2&cursor="+body.Next).Code != 200 || repository.filter.After == nil {
		testingTarget.Fatal("cursor not decoded")
	}
	if request("titles?sort=updated&cursor="+body.Next).Code != 400 {
		testingTarget.Fatal("cross-sort cursor accepted")
	}
	if request("admin/titles").Code != 401 {
		testingTarget.Fatal("anonymous administrator route allowed")
	}
}

func TestCatalogueAgeQueryIncludesUnknownAndPublicDTO(t *testing.T) {
	repo := &catalogueRepository{}
	router := NewRouter(nil)
	NewCatalogHandler(application.New(repo, func() string { return "id" })).Register(router.Group("/api/v1"), NewIdentityHandler(nil, "http://localhost:8080", false))
	for _, raw := range []string{"0", "12", "16", "18"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/titles?age_rating="+raw, nil))
		if response.Code != 200 || repo.filter.AgeRating == nil || strconv.Itoa(*repo.filter.AgeRating) != raw {
			t.Fatal("age filter lost", response.Code)
		}
		var body struct {
			Items []struct {
				AgeRating int `json:"age_rating"`
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || len(body.Items) != 1 || body.Items[0].AgeRating != 0 {
			t.Fatal("unknown classification must not be invented", err)
		}
	}
}
