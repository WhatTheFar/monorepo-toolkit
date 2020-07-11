package interactor_impl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/whatthefar/monorepo-toolkit/pkg/core"
	. "github.com/whatthefar/monorepo-toolkit/pkg/interactor"
)

func NewBuildProjectsInteractor(
	git core.GitGateway,
	pipeline core.PipelineGateway,
	presenter BuildProjectsOutput,
) BuildProjectsInteractor {
	return &buildProjectsInteractor{
		ListChangesInteractor: &listChangesInteractor{git, pipeline},
		presenter:             presenter,
		pipeline:              pipeline,
	}
}

type buildProjectsInteractor struct {
	ListChangesInteractor
	presenter BuildProjectsOutput
	pipeline  core.PipelineGateway
}

func (it *buildProjectsInteractor) BuildPaths(ctx context.Context, paths []string, workflowID string) {
	paths, err := it.ListChanges(ctx, paths, workflowID)
	if err != nil {
		it.presenter.ThrowError(errors.Wrapf(err, `can't list paths with changes for workflow ID "%s"`, workflowID))
		return
	}
	projectNames := it.projectsNameFor(paths)
	it.buildFor(ctx, projectNames)
}

const (
	joinProjectPrefix    = "|"
	joinProjectSeparater = "|"
	joinProjectPostfix   = "|"
)

func (it *buildProjectsInteractor) BuildPathsOnce(ctx context.Context, paths []string, workflowID string) {
	paths, err := it.ListChanges(ctx, paths, workflowID)
	if err != nil {
		it.presenter.ThrowError(errors.Wrapf(err, `can't list paths with changes for workflow ID "%s"`, workflowID))
		return
	}
	projectNames := it.projectsNameFor(paths)
	projectNames = []string{
		fmt.Sprintf(
			"%s%s%s",
			joinProjectPrefix,
			strings.Join(projectNames, joinProjectSeparater),
			joinProjectPostfix,
		),
	}
	it.buildFor(ctx, projectNames)
}

func (it *buildProjectsInteractor) projectsNameFor(paths []string) []string {
	if len(paths) == 0 {
		return []string{}
	}
	projectNames := make([]string, len(paths))
	for i, path := range paths {
		projectName := filepath.Base(path)
		projectNames[i] = projectName
	}
	return projectNames
}

const (
	buildMaxSecondsDefault        = 15*time.Minute + 500*time.Millisecond
	buildCheckAfterSecondsDefault = 15 * time.Second
)

var (
	buildMaxSeconds        = buildMaxSecondsDefault
	buildCheckAfterSeconds = buildCheckAfterSecondsDefault
)

type buildStatus struct {
	projectName string
	buildID     string
	outcome     *string
}

func (it *buildProjectsInteractor) buildFor(ctx context.Context, projectNames []string) {
	statuses := make([]*buildStatus, 0)
	for _, projectName := range projectNames {
		buildID, err := it.pipeline.TriggerBuild(ctx, projectName)
		// TODO: `core` package provide behaviour checking for retrying the request
		if err != nil {
			it.presenter.ThrowError(errors.Wrapf(err, `can't trigger build for project "%s"`, projectName))
			return
		}
		if buildID == nil {
			it.presenter.NoBuildTriggeredFor(projectName)
		} else {
			it.presenter.BuildTriggeredFor(projectName, *buildID)
			status := &buildStatus{
				projectName: projectName,
				buildID:     *buildID,
				outcome:     nil,
			}
			statuses = append(statuses, status)
		}
	}

	if len(statuses) == 0 {
		it.presenter.AllBuildSucceeded(projectNames)
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

	var waitingList []*BuildInfo
loop:
	for {
		waitingList = make([]*BuildInfo, 0)
		for _, s := range statuses {
			if s.outcome == nil {
				outcome, err := it.pipeline.BuildStatus(ctx, s.buildID)
				// TODO: `core` package provide behaviour checking for retrying the request
				if err != nil {
					it.presenter.ThrowError(errors.Wrapf(err, `can't get build status for build ID "%s"`, s.buildID))
					return
				}
				s.outcome = outcome

				if s.outcome != nil {
					switch *s.outcome {
					case "success":
					case "skipped":
						it.presenter.BuildSkippedFor(s.projectName)
					case "failed":
						it.presenter.BuildFailedFor(s.projectName, s.buildID)
						return
					default:
						panic("unknown build status, this should not occur")
					}
				} else {
					info := BuildInfo{ProjectName: s.projectName, BuildID: s.buildID}
					waitingList = append(waitingList, &info)
				}
			}
		}
		if len(waitingList) > 0 {
			it.presenter.WaitingFor(waitingList)
		} else {
			it.presenter.AllBuildSucceeded(projectNames)
			return
		}
		select {
		case <-done:
			break loop
		case <-ticker.C:
			continue
		}
	}

	it.presenter.Timeout(buildMaxSeconds)
	it.presenter.KillingBuilds(waitingList)
	for _, s := range statuses {
		if s.outcome == nil {
			// kill unfinished build
			err := it.pipeline.KillBuild(ctx, s.buildID)
			if err != nil {
				it.presenter.KillBuildError(
					s.projectName,
					errors.Wrapf(err, `can't kill build ID "%s"`, s.buildID),
				)
				// do not return here, try to kill all build
			}
		}
	}
	it.presenter.NotFinishedBuildsKilled()
	return
}
