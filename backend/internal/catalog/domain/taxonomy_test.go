package domain

import "testing"

func TestCollectionsHaveIndependentMetadataAndOrderedUniqueTitles(t *testing.T) {
	item := Collection{Name: "Ғарыш оқиғалары", Slug: "space-stories", Visible: true, TitleIDs: []string{"second", "first"}}
	if !item.Valid() {
		t.Fatal("valid collection rejected")
	}
	item.TitleIDs = []string{"same", "same"}
	if item.Valid() {
		t.Fatal("duplicate titles accepted")
	}
	item.TitleIDs = nil
	item.Position = -1
	if item.Valid() {
		t.Fatal("negative editorial order accepted")
	}
}
