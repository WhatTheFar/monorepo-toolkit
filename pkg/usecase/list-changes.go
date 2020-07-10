//go:generate mockgen -destination mock/list-changes.go . ListChangesUseCase

package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/whatthefar/monorepo-toolkit/pkg/core"
)

type ListChangesUseCase interface {
	ListChanges(ctx context.Context, paths []string, workflowID string) ([]string, error)
}

type listChangesUseCase struct {
	git      core.GitGateway
	pipeline core.PipelineGateway
}

func (uc *listChangesUseCase) ListChanges(ctx context.Context, paths []string, workflowID string) ([]string, error) {
	lastCommit, err := uc.pipeline.LastSuccessfulCommit(ctx, workflowID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get last succesful commit for workflow ID %s", workflowID)
	}
	currentCommit := uc.pipeline.CurrentCommit()
	// Since a local git repository might be a shallow clone,
	// we have to ensure there is enough information for listing changes.
	uc.git.EnsureHavingCommitFromTip(ctx, lastCommit)

	changes, err := uc.git.DiffNameOnly(core.Hash(lastCommit), core.Hash(currentCommit))

	return filterOnlyPathsWithChanges(paths, changes), nil
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
