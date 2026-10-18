package postgres

import "testing"

func TestAuditStatesOnlyExposeSafeScalarChanges(t *testing.T) {
	before, after := auditStates(map[string]any{
		"before":   map[string]any{"role": "reader", "email": "private@example.test", "token": "secret"},
		"after":    map[string]any{"role": "moderator", "hidden": true, "password_hash": "private", "blocked": map[string]any{"token": "secret"}},
		"old_role": map[string]any{"token": "secret"},
	})
	if len(before) != 1 || before["role"] != "reader" || len(after) != 2 || after["role"] != "moderator" || after["hidden"] != true {
		t.Fatal("unsafe or missing projection", before, after)
	}
}

func TestAuditCoverSelectionAndRemovalRetainOnlyAssetIdentity(t *testing.T) {
	for _, next := range []string{"new-cover", ""} {
		before, after := auditStates(map[string]any{
			"before": map[string]any{"cover_asset_id": "old-cover", "object_key": "private-object"},
			"after":  map[string]any{"cover_asset_id": next, "source_key": "private-source"},
		})
		if len(before) != 1 || before["cover_asset_id"] != "old-cover" || len(after) != 1 || after["cover_asset_id"] != next {
			t.Fatal("cover change omitted or private storage data exposed", before, after)
		}
	}
}
