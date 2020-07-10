//go:generate mockgen -destination mock/list-changes.go . ListChangesInteractor

package interactor

import (
	"context"
)

type ListChangesInteractor interface {
	ListChanges(ctx context.Context, paths []string, workflowID string) ([]string, error)
}
