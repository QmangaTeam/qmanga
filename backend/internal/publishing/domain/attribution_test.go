package domain

import "testing"

func TestPreparationMethodDoesNotImplyHumanReview(t *testing.T) {
	for _, method := range []string{"", "unspecified", "manual", "machine_assisted", "machine"} {
		input := Attribution{Source: "Licensed source", Credits: "Translation team", Confirmed: true, PreparationMethod: method}
		if !input.Valid() {
			t.Fatalf("valid method rejected: %s", method)
		}
		if input.HumanReviewed {
			t.Fatal("method invented human review")
		}
	}
	input := Attribution{Source: "Licensed source", Credits: "Translation team", Confirmed: true, PreparationMethod: "certified"}
	if input.Valid() {
		t.Fatal("unsupported preparation method accepted")
	}
}
