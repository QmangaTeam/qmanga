package application

import (
	"context"
	"testing"
)

type statusRepository struct {
	Repository
	actor string
	ids   []string
}

func (repository *statusRepository) Statuses(_ context.Context, actor string, ids []string) (map[string]string, error) {
	repository.actor = actor
	repository.ids = ids
	return map[string]string{"one": "reading"}, nil
}
func TestStatusesBoundAndDeduplicate(t *testing.T) {
	repository := &statusRepository{}
	service := New(repository)
	result, failure := service.Statuses(context.Background(), "owner", []string{"one", "one", "two"})
	if failure != nil || result["one"] != "reading" || repository.actor != "owner" || len(repository.ids) != 2 {
		t.Fatal(result, failure, repository.ids)
	}
	for _, ids := range [][]string{{""}, {string(make([]byte, 129))}, make([]string, 61)} {
		if _, failure = service.Statuses(context.Background(), "owner", ids); failure == nil {
			t.Fatal("invalid batch accepted")
		}
	}
}
