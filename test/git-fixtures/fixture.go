package gitfixture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	rootDir string

	basic      *gitHubRepository = nil
	shallow    *gitHubRepository = nil
	pipeline   *gitHubRepository = nil
	submodules *gitHubRepository = nil
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
			owner:     "WhatTheFar",
			repo:      "monorepo-toolkit-git-fixture-basic",
			relDir:    "test/git-fixtures/basic",
			submodule: "git-fixture-basic",
		}
	}
	return basic
}

func BasicShallowRepository() GitRepository {
	if shallow == nil {
		shallow = &gitHubRepository{
			owner:     "WhatTheFar",
			repo:      "monorepo-toolkit-git-fixture-basic",
			relDir:    "test/git-fixtures/basic-shallow",
			submodule: "git-fixture-basic-shallow",
		}
	}
	return shallow
}

func PipelineRepository() GitRepository {
	if pipeline == nil {
		pipeline = &gitHubRepository{
			owner:     "WhatTheFar",
			repo:      "monorepo-toolkit-git-fixture-pipeline",
			relDir:    "test/git-fixtures/pipeline",
			submodule: "git-fixture-pipeline",
		}
	}
	return pipeline
}

func SubmodulesRepository() GitRepository {
	if submodules == nil {
		submodules = &gitHubRepository{
			owner:     "WhatTheFar",
			repo:      "monorepo-toolkit-git-fixture-submodules",
			relDir:    "test/git-fixtures/submodules",
			submodule: "git-fixture-submodules",
		}
	}
	return submodules
}

type GitRepository interface {
	IsShallow() bool
	Owner() string
	Repository() string
	CommitURL(commit string) string
	CompareURL(from, to string) string
	WorkDir() string
	DotGit() string
	DeleteWorkDir()
	DeleteDotGit()
	SubmoduleUpdate()
}

type gitHubRepository struct {
	owner     string
	repo      string
	relDir    string
	submodule string
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

func (r *gitHubRepository) CommitURL(commit string) string {
	return fmt.Sprintf("%s/commit/%s", r.repositoryURL(), commit)
}

func (r *gitHubRepository) CompareURL(from, to string) string {
	return fmt.Sprintf("%s/compare/%s..%s", r.repositoryURL(), from, to)
}

func (r *gitHubRepository) WorkDir() string {
	return filepath.Join(getRootDir(), r.relDir)
}

func (r *gitHubRepository) DotGit() string {
	return filepath.Join(getRootDir(), ".git", "modules", r.submodule)
}

func (r *gitHubRepository) DeleteWorkDir() {
	err := os.RemoveAll(r.WorkDir())
	if err != nil {
		panic(err)
	}
	return
}

func (r *gitHubRepository) DeleteDotGit() {
	err := os.RemoveAll(r.DotGit())
	if err != nil {
		panic(err)
	}
	return
}

func (r *gitHubRepository) SubmoduleUpdate() {
	cmd := exec.
		Command("git", "submodule", "update", r.relDir)
	cmd.Dir = getRootDir()
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return
}

func (r *gitHubRepository) IsShallow() bool {
	shallowPath := filepath.Join(r.DotGit(), "shallow")
	_, err := os.Stat(shallowPath)
	return !os.IsNotExist(err)
}
