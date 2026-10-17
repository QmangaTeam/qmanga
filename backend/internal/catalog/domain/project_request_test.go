package domain

import (
	"strings"
	"testing"
)

func TestProjectRequestValidation(testingTarget *testing.T) {
	input := ProjectRequestInput{TeamID: "team", Name: "Дала үні", Kind: "manhwa", SourceURL: "https://example.test/story", Note: "Қазақша аударма дайындаймыз"}
	request, failure := NewProjectRequest("request-id", input)
	if failure != nil || request.Name != input.Name || request.Slug == "" || request.Status != "pending" {
		testingTarget.Fatal("valid new-title request rejected", request, failure)
	}
	input.TitleID = "existing"
	input.Name = ""
	input.Kind = ""
	if _, failure = NewProjectRequest("request-id", input); failure != nil {
		testingTarget.Fatal("existing title request rejected", failure)
	}
	for _, bad := range []ProjectRequestInput{{TeamID: "team", Note: input.Note}, {TeamID: "team", TitleID: "existing", Note: "short"}, {TeamID: "team", TitleID: "existing", Note: input.Note, SourceURL: "javascript:alert(1)"}, {TeamID: "team", TitleID: "existing", Note: strings.Repeat("a", 2001)}} {
		if _, failure = NewProjectRequest("request", bad); failure == nil {
			testingTarget.Fatal("invalid request accepted")
		}
	}
	if ValidRequestDecision("reject", "") || ValidRequestDecision("publish", "reason") || !ValidRequestDecision("reject", "Нужно уточнить источник") {
		testingTarget.Fatal("invalid decision validation")
	}
}
