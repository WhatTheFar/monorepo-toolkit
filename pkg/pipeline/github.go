//go:generate mockgen -destination mock/github.go . GitHubActionEnv

package pipeline

import (
	"context"
	"sort"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
)

type GitHubActionEnv interface {
	Token() string
	Ref() string
	Branch() string
	Sha() string
	Owner() string
	Repository() string
}

func NewGitHubActionGateway(ctx context.Context, env GitHubActionEnv) core.PipelineGateway {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: env.Token()},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &gitHubActionGateway{client: client, env: env}
}

type gitHubActionGateway struct {
	client *github.Client
	env    GitHubActionEnv
}

// get hash of last succesfull build commit only commits of 'build' job are considered
func (s *gitHubActionGateway) LastSuccessfulCommit(
	ctx context.Context,
	workflowID string,
) (core.Hash, error) {
	opts := &github.ListWorkflowRunsOptions{
		Branch: s.env.Branch(),
	}
	workflowRuns, _, err := s.client.Actions.ListWorkflowRunsByFileName(
		ctx,
		s.env.Owner(),
		s.env.Repository(),
		workflowID,
		opts,
	)
	if err != nil {
		return "", errors.Wrapf(err, "can't list workflow runs by workflow ID, %s", workflowID)
	}
	// Descending sort workflows by run number
	sort.SliceStable(workflowRuns.WorkflowRuns, func(i, j int) bool {
		return *workflowRuns.WorkflowRuns[i].RunNumber > *workflowRuns.WorkflowRuns[j].RunNumber
	})

	for _, run := range workflowRuns.WorkflowRuns {
		if run.Conclusion != nil && *run.Conclusion == "success" {
			return core.Hash(*run.HeadSHA), nil
		}
	}

	return "", nil
}

// get hash of current commit
func (s *gitHubActionGateway) CurrentCommit() core.Hash {
	return core.Hash(s.env.Sha())
}

// start build of given project
// outputs build request id
func (s *gitHubActionGateway) TriggerBuild(projectName string) (string error) {
	panic("not implemented") // TODO: Implement
}

// get status of build identified by given build number
// outputs one of: success | failed | null
func (s *gitHubActionGateway) BuildStatus(buildID string) (string, error) {
	panic("not implemented") // TODO: Implement
}

// kills running build identified by given build number
func (s *gitHubActionGateway) KillBuild(buildID string) error {
	panic("not implemented") // TODO: Implement
}
