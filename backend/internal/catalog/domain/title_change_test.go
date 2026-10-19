package domain

import "testing"

func TestTitleChangeSnapshotIncludesTaxonomyAndArtwork(t *testing.T) {
	base := ChangeSnapshot{Data: ProposedTitleData{Name: "Title", GenreIDs: []string{"b", "a"}, CategoryIDs: []string{"category"}}, CoverID: "cover"}
	reordered := base
	reordered.Data.GenreIDs = []string{"a", "b"}
	if base.Version() != reordered.Version() {
		t.Fatal("set ordering changed version")
	}
	changed := base
	changed.Data.CategoryIDs = []string{"new"}
	if base.Version() == changed.Version() {
		t.Fatal("category change was missed")
	}
	changed = base
	changed.BannerID = "banner"
	if base.Version() == changed.Version() {
		t.Fatal("banner change was missed")
	}
}
func TestTitleChangePreservesAdministrativeFields(t *testing.T) {
	base := Metadata{Slug: "stable", Name: "Before", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr", Demo: true, Featured: true, Listed: true}
	proposal := ProposedTitleData{Name: "After", Kind: "manhwa", OriginalStatus: "completed", TranslationStatus: "paused", Direction: "rtl", AgeRating: 16, GenreIDs: []string{"genre"}, CategoryIDs: []string{"category"}}
	title, err := proposal.Apply("title", base)
	if err != nil {
		t.Fatal(err)
	}
	if title.Slug != "stable" || !title.Demo || !title.Listed || !title.Featured || title.Name != "After" || *title.AgeRating != 16 {
		t.Fatal("administrative state overwritten", title)
	}
}
func TestChangeSubmissionRequiresReasonAndIndependentEvidence(t *testing.T) {
	input := TitleChangeInput{BaseVersion: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", Reason: "Verified source", Links: []string{"https://example.test/source"}, CoverAction: "keep", BannerAction: "keep"}
	if err := input.Validate(0, false, false); err != nil {
		t.Fatal(err)
	}
	noProof := input
	noProof.Links = nil
	noProof.CoverAction = "replace"
	if noProof.Validate(0, true, false) == nil {
		t.Fatal("proposed artwork alone was treated as evidence")
	}
	bad := input
	bad.Links = []string{"javascript:alert(1)"}
	if bad.Validate(0, false, false) == nil {
		t.Fatal("unsafe link accepted")
	}
	bad = input
	bad.CoverAction = "replace"
	if bad.Validate(0, false, false) == nil {
		t.Fatal("replacement without image accepted")
	}
	bad = input
	bad.Reason = ""
	if bad.Validate(1, false, false) == nil {
		t.Fatal("missing reason accepted")
	}
}
