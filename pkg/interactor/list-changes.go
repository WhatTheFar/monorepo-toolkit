//go:generate mockgen -destination mock/list-changes.go . ListChangesInteractor

package interactor

import (
	"context"
)

type ListChangesInteractor interface {
	ListChanges(ctx context.Context, paths []string, workflowID string) (changedPaths []string, err error)
	ListProjects(ctx context.Context, paths []string, workflowID string) (projectNames []string, err error)
	ListProjectsJoined(ctx context.Context, paths []string, workflowID string) (projectName string, err error)
}
