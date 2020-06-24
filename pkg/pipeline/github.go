//go:generate mockgen -destination mock/github.go . GitHubActionEnv

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
)

const (
	triggerBuildWaitForSecond = 5
)

type GitHubActionEnv interface {
	Token() string
	Ref() string
	Branch() string
	Sha() string
	Owner() string
	Repository() string
	EventType() string
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
func (s *gitHubActionGateway) TriggerBuild(ctx context.Context, projectName string) (string, error) {
	eventType := s.env.EventType()
	if eventType == "" {
		eventType = fmt.Sprintf("build-%s", projectName)
	}
	payload := json.RawMessage(fmt.Sprintf(`{ "job": "%s" }`, projectName))
	opts := github.DispatchRequestOptions{
		EventType:     eventType,
		ClientPayload: &payload,
	}
	now := time.Now()
	_, _, err := s.client.Repositories.Dispatch(ctx, s.env.Owner(), s.env.Repository(), opts)
	if err != nil {
		return "", errors.Wrap(err, "can't dispatch event")
	}
	id, err := getLastRepositoryDispatchRunID(ctx, s, projectName, now)
	if id == 0 {
		return "", err
	}
	return fmt.Sprintf("%d", id), err
}

func getLastRepositoryDispatchRunID(
	ctx context.Context,
	s *gitHubActionGateway,
	projectName string,
	now time.Time,
) (int64, error) {
	owner, repo := s.env.Owner(), s.env.Repository()
	for i := 0; i < triggerBuildWaitForSecond; i++ {
		opts := &github.ListWorkflowRunsOptions{
			Event: "repository_dispatch",
		}
		workflowRuns, _, err := s.client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
		if err != nil {
			return 0, errors.Wrap(err, "can't list workflow runs for a repository")
		}

		for _, run := range workflowRuns.WorkflowRuns {
			if run.GetCreatedAt().Before(now) {
				continue
			}
			runID := run.GetID()
			jobs, _, err := s.client.Actions.ListWorkflowJobs(ctx, owner, repo, runID, nil)
			if err != nil {
				return 0, errors.Wrapf(err, "can't list jobs of a workflow run, ID %d", runID)
			}
			for _, job := range jobs.Jobs {
				re := regexp.MustCompile(fmt.Sprintf(`%s`, projectName))
				if re.MatchString(job.GetName()) == true {
					return runID, nil
				}
			}
		}

		time.Sleep(time.Second)
	}
	return 0, nil
}

func getRunIDFromJobURL(url string) (int64, error) {
	re := regexp.MustCompile(`^https://api\.github\.com/repos/[^/ ]+/[^/ ]+/actions/runs/(?P<id>\d+)/jobs$`)
	match := re.FindStringSubmatch(url)
	if len(match) == 0 {
		return 0, errors.Errorf("invalid job url: %s", url)
	}
	for i, name := range re.SubexpNames() {
		if name == "id" {
			id, err := strconv.ParseInt(match[i], 10, 64)
			if err != nil {
				return 0, errors.Wrapf(err, "invalid job ID: %s", match[i])
			}
			return id, nil
		}
	}
	return 0, nil
}

// get status of build identified by given build number
// outputs one of: success | failed | null
func (s *gitHubActionGateway) BuildStatus(ctx context.Context, buildID string) (string, error) {
	runID, err := strconv.ParseInt(buildID, 10, 64)
	if err != nil {
		return "", errors.Wrapf(err, "invalid build ID: %s", buildID)
	}
	workflowRun, _, err := s.client.Actions.GetWorkflowRunByID(ctx, s.env.Owner(), s.env.Repository(), runID)
	if err != nil {
		return "", errors.Wrapf(err, "can't get a workflow run, ID %d", runID)
	}
	switch workflowRun.GetConclusion() {
	case "success":
		return "success", nil
	case "failure", "cancelled":
		return "failed", nil
	default:
		return "", nil
	}
}

// kills running build identified by given build number
func (s *gitHubActionGateway) KillBuild(buildID string) error {
	panic("not implemented") // TODO: Implement
}
