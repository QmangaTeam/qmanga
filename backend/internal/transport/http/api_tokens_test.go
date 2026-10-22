package http

import "testing"

func TestAPITokenRoutesDenyCredentialAndUnrelatedAdministration(t *testing.T) {
	for _, path := range []string{"/api/v1/admin/api-tokens", "/api/v1/auth/csrf", "/api/v1/admin/audit", "/api/v1/moderation/accounts", "/api/v1/teams/x", "/api/v1/reading/library"} {
		if apiTokenRouteAllowed("GET", path) {
			t.Errorf("allowed %s", path)
		}
	}
	for _, pair := range [][2]string{{"GET", "/api/v1/auth/me"}, {"POST", "/api/v1/admin/titles"}, {"PUT", "/api/v1/studio/revisions/:id/pages"}, {"POST", "/api/v1/studio/revisions/:id/publish"}} {
		if !apiTokenRouteAllowed(pair[0], pair[1]) {
			t.Errorf("denied %v", pair)
		}
	}
}
