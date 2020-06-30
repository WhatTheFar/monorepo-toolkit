//go:generate mockgen -destination mock/build-projects.go . BuildProjectsPresenter

package usecase

import (
	"context"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/whatthefar/monorepo-toolkit/pkg/core"
)

type BuildProjectsUseCase interface {
	BuildFor(ctx context.Context, paths []string, workflowID string)
}

type BuildProjectsPresenter interface {
	BuildTriggeredFor(projectName string)
	NoBuildTriggeredFor(projectName string)
	BuildFailedFor(projectName string)
	BuildSkippedFor(projectName string)
	WaitingFor(projectNames []string)
	AllBuildSucceeded()
	Timeout()
	KillingBuildsFor(projectNames []string)
	KillBuildErrorFor(err error, projectName string)
	NotFinishedBuildsKilled()

	ThrowError(err error)
}

type iListProjects interface {
	projectsFor(paths []string) []string
}

type buildProjectsUseCase struct {
	ListChangesUseCase
	iListProjects
	presenter BuildProjectsPresenter
	pipeline  core.PipelineGateway
}

type buildStatus struct {
	projectName string
	buildID     string
	outcome     *string
}

const (
	buildMaxSecondsDefault        = 15*time.Minute + 500*time.Millisecond
	buildCheckAfterSecondsDefault = 15 * time.Second
)

var (
	buildMaxSeconds        = buildMaxSecondsDefault
	buildCheckAfterSeconds = buildCheckAfterSecondsDefault
)

func (uc *buildProjectsUseCase) BuildFor(ctx context.Context, paths []string, workflowID string) {
	paths, err := uc.ListChanges(ctx, paths, workflowID)
	if err != nil {
		uc.presenter.ThrowError(errors.Wrapf(err, `can't list paths with changes for workflow ID "%s"`, workflowID))
		return
	}
	projectNames := uc.projectsFor(paths)

	statuses := make([]*buildStatus, 0)
	for _, projectName := range projectNames {
		buildID, err := uc.pipeline.TriggerBuild(ctx, projectName)
		// TODO: `core` package provide behaviour checking for retrying the request
		if err != nil {
			uc.presenter.ThrowError(errors.Wrapf(err, `can't trigger build for project "%s"`, projectName))
			return
		}
		if buildID == nil {
			uc.presenter.NoBuildTriggeredFor(projectName)
		} else {
			uc.presenter.BuildTriggeredFor(projectName)
			status := &buildStatus{
				projectName: projectName,
				buildID:     *buildID,
				outcome:     nil,
			}
			statuses = append(statuses, status)
		}
	}

	if len(statuses) == 0 {
		uc.presenter.AllBuildSucceeded()
		return
	}

	ticker := time.NewTicker(buildCheckAfterSeconds)
	done := make(chan bool)

	defer func() {
		ticker.Stop()
		close(done)
	}()

	go func() {
		select {
		case <-done:
			return
		case <-time.After(buildMaxSeconds):
			done <- true
		}
	}()

	var waitingList []string
loop:
	for {
		waitingList = make([]string, 0)
		for _, s := range statuses {
			if s.outcome == nil {
				outcome, err := uc.pipeline.BuildStatus(ctx, s.buildID)
				// TODO: `core` package provide behaviour checking for retrying the request
				if err != nil {
					uc.presenter.ThrowError(errors.Wrapf(err, `can't get build status for build ID "%s"`, s.buildID))
					return
				}
				s.outcome = outcome

				if s.outcome != nil {
					switch *s.outcome {
					case "success":
					case "skipped":
						uc.presenter.BuildSkippedFor(s.projectName)
					case "failed":
						uc.presenter.BuildFailedFor(s.projectName)
						return
					default:
						panic("unknown build status, this should not occur")
					}
				} else {
					waitingList = append(waitingList, s.projectName)
				}
			}
		}
		if len(waitingList) > 0 {
			uc.presenter.WaitingFor(waitingList)
		} else {
			uc.presenter.AllBuildSucceeded()
			return
		}
		select {
		case <-done:
			break loop
		case <-ticker.C:
			continue
		}
	}

	uc.presenter.Timeout()
	uc.presenter.KillingBuildsFor(waitingList)
	for _, s := range statuses {
		if s.outcome == nil {
			// kill unfinished build
			err := uc.pipeline.KillBuild(ctx, s.buildID)
			if err != nil {
				uc.presenter.KillBuildErrorFor(errors.Wrapf(err, `can't kill build ID "%s"`, s.buildID), s.projectName)
				// do not return here, try to kill all build
			}
		}
	}
	uc.presenter.NotFinishedBuildsKilled()
	return
}
