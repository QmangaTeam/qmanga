package readiness

import (
	"context"
	"time"
)

type Pinger interface {
	Ping(contextInstance context.Context) error
}

type HealthReport struct {
	Status   string
	Version  string
	Postgres string
	S3       string
}

type CheckHealthUseCase struct {
	version        string
	postgresPinger Pinger
	s3Pinger       Pinger
	timeout        time.Duration
}

func NewCheckHealthUseCase(version string, postgresPinger Pinger, s3Pinger Pinger) *CheckHealthUseCase {
	return &CheckHealthUseCase{
		version:        version,
		postgresPinger: postgresPinger,
		s3Pinger:       s3Pinger,
		timeout:        2 * time.Second,
	}
}

func (useCase *CheckHealthUseCase) Execute(parentContext context.Context) (HealthReport, bool) {
	checkContext, cancelFunc := context.WithTimeout(parentContext, useCase.timeout)
	defer cancelFunc()

	postgresStatus := "ok"
	if useCase.postgresPinger == nil {
		postgresStatus = "unavailable"
	} else if pingError := useCase.postgresPinger.Ping(checkContext); pingError != nil {
		postgresStatus = "unavailable"
	}

	s3Status := "ok"
	if useCase.s3Pinger == nil {
		s3Status = "unavailable"
	} else if pingError := useCase.s3Pinger.Ping(checkContext); pingError != nil {
		s3Status = "unavailable"
	}

	isHealthy := postgresStatus == "ok" && s3Status == "ok"
	overallStatus := "ok"
	if !isHealthy {
		overallStatus = "degraded"
	}

	report := HealthReport{
		Status:   overallStatus,
		Version:  useCase.version,
		Postgres: postgresStatus,
		S3:       s3Status,
	}

	return report, isHealthy
}
