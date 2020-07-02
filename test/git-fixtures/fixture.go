package gitfixture

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	rootDir string

	basic    *gitHubRepository = nil
	shallow  *gitHubRepository = nil
	pipeline *gitHubRepository = nil
)

func getRootDir() string {
	if rootDir == "" {
		rootDirBytes, err := exec.
			Command("git", "rev-parse", "--show-toplevel").
			Output()
		if err != nil {
			panic(err)
		}
		rootDir = strings.TrimSuffix(string(rootDirBytes), "\n")
	}
	return rootDir
}

func BasicRepository() GitRepository {
	if basic == nil {
		basic = &gitHubRepository{
			owner:  "WhatTheFar",
			repo:   "monorepo-toolkit-git-fixture-basic",
			relDir: "test/git-fixtures/basic",
		}
	}
	return basic
}

func BasicShallowRepository() GitRepository {
	if shallow == nil {
		shallow = &gitHubRepository{
			owner:  "WhatTheFar",
			repo:   "monorepo-toolkit-git-fixture-basic",
			relDir: "test/git-fixtures/basic-shallow",
		}
	}
	return basic
}

func PipelineRepository() GitRepository {
	if pipeline == nil {
		pipeline = &gitHubRepository{
			owner:  "WhatTheFar",
			repo:   "monorepo-toolkit-git-fixture-pipeline",
			relDir: "test/git-fixtures/pipeline",
		}
	}
	return pipeline
}

type GitRepository interface {
	Owner() string
	Repository() string
	CompareURL(from, to string) string
	WorkDir() string
}

type gitHubRepository struct {
	owner  string
	repo   string
	relDir string
}

func (r *gitHubRepository) repositoryURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.owner, r.repo)
}

func (r *gitHubRepository) Owner() string {
	return r.owner
}

func (r *gitHubRepository) Repository() string {
	return r.repo
}

func (r *gitHubRepository) CompareURL(from, to string) string {
	return fmt.Sprintf("%s/compare/%s..%s", r.repositoryURL(), from, to)
}

func (r *gitHubRepository) WorkDir() string {
	return filepath.Join(getRootDir(), r.relDir)
}
