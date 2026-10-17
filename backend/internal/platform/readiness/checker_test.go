package readiness

import (
	"context"
	"errors"
	"testing"
)

type mockPinger struct {
	err error
}

func (pinger *mockPinger) Ping(contextInstance context.Context) error {
	return pinger.err
}

func TestCheckHealthUseCase_AllHealthy(testingTarget *testing.T) {
	postgresMock := &mockPinger{err: nil}
	s3Mock := &mockPinger{err: nil}

	useCase := NewCheckHealthUseCase("0.1.0", postgresMock, s3Mock)
	report, isHealthy := useCase.Execute(context.Background())

	if !isHealthy {
		testingTarget.Fatal("expected isHealthy true, got false")
	}
	if report.Status != "ok" {
		testingTarget.Errorf("expected status ok, got %s", report.Status)
	}
	if report.Postgres != "ok" {
		testingTarget.Errorf("expected postgres ok, got %s", report.Postgres)
	}
	if report.S3 != "ok" {
		testingTarget.Errorf("expected s3 ok, got %s", report.S3)
	}
	if report.Version != "0.1.0" {
		testingTarget.Errorf("expected version 0.1.0, got %s", report.Version)
	}
}

func TestCheckHealthUseCase_PostgresDegraded(testingTarget *testing.T) {
	postgresMock := &mockPinger{err: errors.New("sensitive connection string failed")}
	s3Mock := &mockPinger{err: nil}

	useCase := NewCheckHealthUseCase("0.1.0", postgresMock, s3Mock)
	report, isHealthy := useCase.Execute(context.Background())

	if isHealthy {
		testingTarget.Fatal("expected isHealthy false, got true")
	}
	if report.Status != "degraded" {
		testingTarget.Errorf("expected status degraded, got %s", report.Status)
	}
	if report.Postgres != "unavailable" {
		testingTarget.Errorf("expected postgres unavailable without leaking raw error, got %s", report.Postgres)
	}
	if report.S3 != "ok" {
		testingTarget.Errorf("expected s3 ok, got %s", report.S3)
	}
}

func TestCheckHealthUseCase_NilPingersUnhealthy(testingTarget *testing.T) {
	useCase := NewCheckHealthUseCase("0.1.0", nil, nil)
	report, isHealthy := useCase.Execute(context.Background())

	if isHealthy {
		testingTarget.Fatal("expected isHealthy false for nil pingers, got true")
	}
	if report.Status != "degraded" {
		testingTarget.Errorf("expected status degraded, got %s", report.Status)
	}
	if report.Postgres != "unavailable" {
		testingTarget.Errorf("expected postgres unavailable, got %s", report.Postgres)
	}
	if report.S3 != "unavailable" {
		testingTarget.Errorf("expected s3 unavailable, got %s", report.S3)
	}
}
