package domain

import "testing"

func TestPublicationMetadata(t *testing.T) {
	base := Metadata{Slug: "test", Name: "Test", Kind: "manhwa", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"}
	base.Publication = Publication{Year: 2023, Publishers: []string{"Publisher"}, Formats: []string{"Webtoon"}, Country: "Корея", Source: "Новелла", Links: []ExternalLink{{Label: "Official", URL: "https://example.com/title"}}}
	if _, err := NewTitle("title", base); err != nil {
		t.Fatal(err)
	}
	for _, url := range []string{"javascript:alert(1)", "data:text/html,test", "https://user:password@example.com", "/relative"} {
		bad := base
		bad.Publication.Links = []ExternalLink{{Label: "Official", URL: url}}
		if _, err := NewTitle("title", bad); err == nil {
			t.Fatalf("accepted unsafe external link: %s", url)
		}
	}
	base.Publication.Year = 99999
	if _, err := NewTitle("title", base); err == nil {
		t.Fatal("accepted invalid publication year")
	}
}
