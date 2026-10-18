package domain

import "unicode/utf8"

type Collection struct {
	ID, Name, Slug, Description string
	Position                    int
	Visible                     bool
	TitleIDs                    []string
}

func (collection Collection) Valid() bool {
	if !ValidGenre(collection.Name, collection.Slug) || utf8.RuneCountInString(collection.Description) > 4000 || collection.Position < 0 || collection.Position > 1000000 || len(collection.TitleIDs) > 500 {
		return false
	}
	return ValidReferences(collection.TitleIDs, 500)
}
func ValidReferences(ids []string, maximum int) bool {
	if len(ids) > maximum {
		return false
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || len(id) > 128 || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
