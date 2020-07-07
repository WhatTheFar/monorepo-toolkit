package git

import (
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
)

type gitGateway struct {
	repo *gogit.Repository
}

func NewGitGateway(path string) (core.GitGateway, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		if err == gogit.ErrRepositoryNotExists {
			return nil, errors.Wrap(err, "can't open a repository")
		}
		// unknown error
		return nil, errors.Wrap(err, "can't open a repository")
	}
	return &gitGateway{repo: repo}, nil
}

func (g *gitGateway) hasCommit(sha core.Hash) (bool, error) {
	_, err := g.repo.CommitObject(plumbing.NewHash(string(sha)))
	if err != nil {
		if err == plumbing.ErrObjectNotFound {
			return false, nil
		}
		return false, errors.Wrapf(err, "can't get a commit object of %s", string(sha))
	}
	return true, nil
}

func (g *gitGateway) DiffNameOnly(from core.Hash, to core.Hash) ([]string, error) {
	fromHash := plumbing.NewHash(string(from))
	toHash := plumbing.NewHash(string(to))

	// get commit objects
	fromCommit, err := g.repo.CommitObject(fromHash)
	if err != nil {
		return nil, errors.Wrap(err, "can't get commit object of the `from` hash")
	}
	toCommit, err := g.repo.CommitObject(toHash)
	if err != nil {
		return nil, errors.Wrap(err, "can't get commit object of the `to` hash")
	}

	// get tree objects
	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "can't get tree object of the `from` commit object")
	}
	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "can't get tree object of the `to` commit object")
	}

	// get changes
	changes, err := object.DiffTree(fromTree, toTree)
	if err != nil {
		return nil, errors.Wrap(err, "can't get diff changes between trees")
	}

	// get file patch from changes
	patch, err := changes.Patch()
	if err != nil {
		return nil, errors.Wrap(err, "can't get file patch from changes")
	}

	changeStats := patch.Stats()

	fileNames := make([]string, len(changeStats))
	for i, v := range changeStats {
		fileNames[i] = v.Name

	}

	return fileNames, nil
}
