package domain

import "testing"

func TestKazakhMetadataAndChapterLabels(testingTarget *testing.T) {
	title, failure := NewTitle("title", Metadata{Slug: "aspan-balalary", Name: "Аспан балалары", OriginalName: "Children of the sky", Aliases: []string{"Аспан", "Sky children"}, Kind: "manhwa", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"})
	if failure != nil || title.Name != "Аспан балалары" || len(title.Aliases) != 2 {
		testingTarget.Fatalf("Kazakh title rejected: %v", failure)
	}
	for _, label := range []string{"1", "1.1", "1.10", "2", "Пролог"} {
		if failure = ValidateChapter(label, 10); failure != nil {
			testingTarget.Fatal(failure)
		}
	}
	if ValidateChapter("", 0) == nil {
		testingTarget.Fatal("invalid chapter accepted")
	}
	for _, metadata := range []Metadata{
		{Slug: "bad/slug", Name: "Аспан", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"},
		{Slug: "valid", Name: "", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"},
		{Slug: "valid", Name: "Аспан", Kind: "video", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"},
	} {
		if _, failure = NewTitle("title", metadata); failure == nil {
			testingTarget.Fatal("invalid metadata accepted")
		}
	}
}

func TestAgeRatingDoesNotAssumeAClassification(t *testing.T) {
	base := Metadata{Slug: "age-test", Name: "Age test", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"}
	if title, err := NewTitle("id", base); err != nil || title.AgeRating != nil {
		t.Fatal("legacy metadata must remain unspecified", err)
	}
	for _, age := range []int{0, 12, 16, 18, -1, 1, 14, 19, 99} {
		base.AgeRating = &age
		_, err := NewTitle("id", base)
		valid := age == 0 || age == 12 || age == 16 || age == 18
		if (err == nil) != valid {
			t.Fatalf("age %d: %v", age, err)
		}
	}
}
