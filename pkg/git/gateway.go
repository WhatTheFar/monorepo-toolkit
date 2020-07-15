package git

import (
	"context"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/pkg/errors"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
)

type gitGateway struct {
	repo    *gogit.Repository
	workDir string
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
	return &gitGateway{repo: repo, workDir: path}, nil
}

var (
	ErrNoCommit = errors.New("no commit found")
)

func (g *gitGateway) EnsureHavingCommitFromTip(ctx context.Context, sha core.Hash) error {
	hasCommit, err := g.hasCommit(sha)
	if err != nil {
		return errors.Wrapf(err, `can't check is there a commit "%s"`, sha)
	}
	if hasCommit == true {
		return nil
	}

	shallows, err := g.repo.Storer.Shallow()
	if shallows != nil {
		remoteName := "origin"
		remote, err := g.repo.Remote(remoteName)
		if err != nil {
			return errors.Wrapf(err, `can't remote "%s"`, remoteName)
		}
		url := remote.Config().URLs[0]
		// TODO: implement auth for private repo
		dotGit := filepath.Join(g.workDir, ".git")
		err = os.RemoveAll(dotGit)
		if err != nil {
			return errors.Wrapf(err, `can't delete .git at "%s"`, dotGit)
		}
		newRepo, err := gogit.PlainCloneContext(ctx, dotGit, true, &gogit.CloneOptions{URL: url})
		if err != nil {
			return errors.Wrapf(err, `can't clone .git for URL "%s"`, url)
		}
		g.repo = newRepo
	} else {
		err := g.repo.FetchContext(ctx, &gogit.FetchOptions{})
		if err != nil {
			return errors.Wrap(err, "can't fetch")
		}
	}

	hasCommit, err = g.hasCommit(sha)
	if err != nil {
		return errors.Wrapf(err, `can't check is there a commit "%s"`, sha)
	}
	if hasCommit != true {
		return errors.Wrapf(ErrNoCommit, `commit "%s" not found`, sha)
	}

	return nil
}

func (g *gitGateway) IsNoCommit(err error) bool {
	if errors.Cause(err) == ErrNoCommit {
		return true
	}
	return false
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

	changedPaths := make([]string, 0)
	for _, v := range changes {
		action, _ := v.Action()
		if action == merkletrie.Delete || action == merkletrie.Modify {
			changedPaths = append(changedPaths, v.From.Name)
		}
		if action == merkletrie.Insert {
			changedPaths = append(changedPaths, v.To.Name)
		}
	}
	return changedPaths, nil
}
