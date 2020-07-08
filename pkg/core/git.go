//go:generate mockgen -destination mock/git.go . GitGateway

package core

import (
	"context"
)

// Hash SHA1 hashed content
type Hash string

type GitGateway interface {
	DiffNameOnly(from, to Hash) ([]string, error)
	EnsureHavingCommitFromTip(ctx context.Context, sha Hash) error
	IsNoCommit(err error) bool
}
