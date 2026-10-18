package application

import (
	"context"
	"testing"
)

type fakeRepository struct{ called bool }

func (r *fakeRepository) Report(context.Context, string, bool, int) (Report, error) {
	r.called = true
	return Report{}, nil
}
func TestPeriodsAreBounded(t *testing.T) {
	for _, days := range []int{0, -1, 1, 6, 8, 29, 31, 91, 100000} {
		repo := &fakeRepository{}
		_, err := New(repo).Report(context.Background(), "actor", true, days)
		if err != ErrInvalid || repo.called {
			t.Fatal(days, err)
		}
	}
	for _, days := range []int{7, 30, 90} {
		repo := &fakeRepository{}
		_, err := New(repo).Report(context.Background(), "actor", true, days)
		if err != nil || !repo.called {
			t.Fatal(days, err)
		}
	}
}
