package interactor_impl

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/whatthefar/monorepo-toolkit/pkg/core"
	. "github.com/whatthefar/monorepo-toolkit/pkg/interactor"
)

func NewListChangesInteractor(git core.GitGateway, pipeline core.PipelineGateway) ListChangesInteractor {
	return &listChangesInteractor{git: git, pipeline: pipeline}
}

type listChangesInteractor struct {
	git      core.GitGateway
	pipeline core.PipelineGateway
}

func (it *listChangesInteractor) ListChanges(ctx context.Context, paths []string, workflowID string) ([]string, error) {
	lastCommit, err := it.pipeline.LastSuccessfulCommit(ctx, workflowID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get last succesful commit for workflow ID %s", workflowID)
	}
	currentCommit := it.pipeline.CurrentCommit()
	// Since a local git repository might be a shallow clone,
	// we have to ensure there is enough information for listing changes.
	it.git.EnsureHavingCommitFromTip(ctx, lastCommit)

	changes, err := it.git.DiffNameOnly(core.Hash(lastCommit), core.Hash(currentCommit))

	return filterOnlyPathsWithChanges(paths, changes), nil
}

func (it *listChangesInteractor) ListProjects(ctx context.Context, paths []string, workflowID string) ([]string, error) {
	changedPaths, err := it.ListChanges(ctx, paths, workflowID)
	if err != nil {
		return nil, errors.Wrapf(err, `can't list paths with changes for workflow ID "%s"`, workflowID)
	}
	projectNames := projectsNameFor(changedPaths)
	return projectNames, nil
}

const (
	joinProjectPrefix    = "|"
	joinProjectSeparater = "|"
	joinProjectPostfix   = "|"
)

func (it *listChangesInteractor) ListProjectsJoined(ctx context.Context, paths []string, workflowID string) (string, error) {
	changedPaths, err := it.ListChanges(ctx, paths, workflowID)
	if err != nil {
		return "", errors.Wrapf(err, `can't list paths with changes for workflow ID "%s"`, workflowID)
	}
	projectNames := projectsNameFor(changedPaths)
	projectNamesJoined := fmt.Sprintf(
		"%s%s%s",
		joinProjectPrefix,
		strings.Join(projectNames, joinProjectSeparater),
		joinProjectPostfix,
	)
	return projectNamesJoined, nil
}

func filterOnlyPathsWithChanges(paths []string, changes []string) []string {
	changesJoined := strings.Join(changes, "\n")
	pathsWithChanges := make([]string, 0)
	for _, path := range paths {
		re := regexp.MustCompile(fmt.Sprintf(`(?m)^%s`, path))
		if re.MatchString(changesJoined) == true {
			pathsWithChanges = append(pathsWithChanges, path)
		}
	}
	return pathsWithChanges
}

func projectsNameFor(paths []string) []string {
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
