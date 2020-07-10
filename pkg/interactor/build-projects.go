//go:generate mockgen -destination mock/build-projects.go . BuildProjectsOutput

package interactor

import (
	"context"
	"time"
)

type BuildProjectsInteractor interface {
	BuildFor(ctx context.Context, paths []string, workflowID string)
}

type BuildInfo struct {
	ProjectName string
	BuildID     string
}

type BuildProjectsOutput interface {
	BuildTriggeredFor(projectName string, buildID string)
	NoBuildTriggeredFor(projectName string)
	BuildFailedFor(projectName string, buildID string)
	BuildSkippedFor(projectName string)
	WaitingFor(buildInfos []*BuildInfo)
	AllBuildSucceeded(projectNames []string)
	Timeout(waitingTime time.Duration)
	KillingBuilds(buildInfos []*BuildInfo)
	KillBuildError(projectName string, err error)
	NotFinishedBuildsKilled()

	ThrowError(err error)
}
