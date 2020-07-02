package gitfixture

import "fmt"

var (
	basic    *gitHubRepository = nil
	pipeline *gitHubRepository = nil
)

func BasicRepository() GitRepository {
	if basic == nil {
		basic = &gitHubRepository{
			owner: "WhatTheFar",
			repo:  "monorepo-toolkit-git-fixture-basic",
		}
	}
	return basic
}

func BasicShallowRepository() GitRepository {
	if shallow == nil {
		shallow = &gitHubRepository{
			owner:  "WhatTheFar",
			repo:   "monorepo-toolkit-git-fixture-basic",
		}
	}
	return basic
}

func PipelineRepository() GitRepository {
	if pipeline == nil {
		pipeline = &gitHubRepository{
			owner: "WhatTheFar",
			repo:  "monorepo-toolkit-git-fixture-pipeline",
		}
	}
	return pipeline
}

type GitRepository interface {
	Owner() string
	Repository() string
	CompareURL(from, to string) string
}

type gitHubRepository struct {
	owner string
	repo  string
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
