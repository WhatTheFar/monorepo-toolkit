//go:generate mockgen -destination mock/ci.go . CI

package controller

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/whatthefar/monorepo-toolkit/pkg/interactor"
)

type CI interface {
	Build(ctx context.Context, paths []string, workflowID string) error
	BuildOnce(ctx context.Context, paths []string, workflowID string) error

	ListProjects(ctx context.Context, paths []string, workflowID string) ([]string, error)
	ListProjectsJoined(ctx context.Context, paths []string, workflowID string) (string, error)
}

func NewCIController(
	listChangeIt interactor.ListChangesInteractor,
	buildProjectsIt interactor.BuildProjectsInteractor,
) CI {
	return &ci{
		ListChangesInteractor:   listChangeIt,
		BuildProjectsInteractor: buildProjectsIt,
	}
}

type ci struct {
	interactor.ListChangesInteractor
	interactor.BuildProjectsInteractor
}

func (c *ci) Build(ctx context.Context, paths []string, workflowID string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "can't get current directory")
	}

	paths = relPathsIfPossible(workDir, paths)
	c.BuildPaths(ctx, paths, workflowID)
	return nil
}
func (c *ci) BuildOnce(ctx context.Context, paths []string, workflowID string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "can't get current directory")
	}

	paths = relPathsIfPossible(workDir, paths)
	c.BuildPaths(ctx, paths, workflowID)
	return nil
}

func (c *ci) ListProjects(ctx context.Context, paths []string, workflowID string) ([]string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "can't get current directory")
	}

	paths = relPathsIfPossible(workDir, paths)
	projects, err := c.ListChangesInteractor.ListProjects(ctx, paths, workflowID)
	if err != nil {
		return nil, errors.Wrap(err, "can't list projects that have changes")
	}
	return projects, nil
}

func (c *ci) ListProjectsJoined(ctx context.Context, paths []string, workflowID string) (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "can't get current directory")
	}

	paths = relPathsIfPossible(workDir, paths)
	project, err := c.ListChangesInteractor.ListProjectsJoined(ctx, paths, workflowID)
	if err != nil {
		return "", errors.Wrap(err, "can't list projects that have changes")
	}
	return project, nil
}

func relPathsIfPossible(workDir string, paths []string) []string {
	for i, path := range paths {
		if filepath.IsAbs(path) {
			rel, err := filepath.Rel(workDir, path)
			if err != nil {
				// An error is returned if targpath can't be made relative to basepath
				// Do noting
				continue
			}
			paths[i] = rel
		}
	}
	return paths
}
