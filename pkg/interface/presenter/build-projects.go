package presenter

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/whatthefar/monorepo-toolkit/pkg/interactor"
)

func NewBuildProjectsPresenter(writer io.Writer) interactor.BuildProjectsOutput {
	return &buildProjectsPresenter{writer: writer}
}

type buildProjectsPresenter struct {
	writer io.Writer
}

func (p *buildProjectsPresenter) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(p.writer, a...)
}

func (p *buildProjectsPresenter) BuildTriggeredFor(projectName string, buildID string) {
	p.Println(
		fmt.Sprintf("Build triggered for project '%s' with number '%s'", projectName, buildID),
	)
}

func (p *buildProjectsPresenter) NoBuildTriggeredFor(projectName string) {
	// TODO: add yellow color to "WARN"
	p.Println(
		fmt.Sprintf("WARN: No build triggered for project '%s'.", projectName),
		"Please check if pipeline is defined in your build tool.",
	)
}

func (p *buildProjectsPresenter) BuildFailedFor(projectName string, buildID string) {
	p.Println(fmt.Sprintf("Build failed for project '%s(%s)'", projectName, buildID))
}

func (p *buildProjectsPresenter) BuildSkippedFor(projectName string) {
	// TODO: add yellow color to "WARN"
	p.Println(
		fmt.Sprintf("WARN: Build was skipped for project '%s'.", projectName),
		"Please check if pipeline is defined in your build tool.",
	)
}

func (p *buildProjectsPresenter) WaitingFor(buildInfos []*interactor.BuildInfo) {
	buildStrs := make([]string, len(buildInfos))
	for i, v := range buildInfos {
		buildStrs[i] = fmt.Sprintf("%s(%s)", v.ProjectName, v.BuildID)

	}
	p.Println(
		fmt.Sprintf("Waiting for build %s...", strings.Join(buildStrs, " ")),
	)
}

func (p *buildProjectsPresenter) AllBuildSucceeded(projectNames []string) {
	p.Println(
		fmt.Sprintf("Build successful for all projects: %s", strings.Join(projectNames, " ")),
	)
}

func (p *buildProjectsPresenter) Timeout(waitingTime time.Duration) {
	p.Println(
		fmt.Sprintf(
			"Timeout! Some builds were not finished withing %d seconds.",
			waitingTime,
		),
	)
}

func (p *buildProjectsPresenter) KillingBuilds(buildInfos []*interactor.BuildInfo) {
	buildStrs := make([]string, len(buildInfos))
	for i, v := range buildInfos {
		buildStrs[i] = fmt.Sprintf("%s(%s)", v.ProjectName, v.BuildID)

	}
	p.Println(
		fmt.Sprintf("Killing not finished builds: %s...", strings.Join(buildStrs, " ")),
	)
}

func (p *buildProjectsPresenter) KillBuildError(projectName string, err error) {
	p.Println(
		fmt.Sprintf("Killing build '%s' error: %s...", projectName, err),
	)
}

func (p *buildProjectsPresenter) NotFinishedBuildsKilled() {
	p.Println("All not finished builds were killed")
}

func (p *buildProjectsPresenter) ThrowError(err error) {
	panic(err)
}
