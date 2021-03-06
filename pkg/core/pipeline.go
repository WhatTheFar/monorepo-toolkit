//go:generate mockgen -destination mock/pipeline.go . PipelineGateway

package core

import (
	"context"
)

type PipelineGateway interface {
	// get hash of last succesfull build commit only commits of 'build' job are considered
	LastSuccessfulCommit(ctx context.Context, workflowID string) (Hash, error)
	// get hash of current commit
	CurrentCommit() Hash

	// start build of given project
	// outputs build request id
	TriggerBuild(ctx context.Context, projectName string) (*string, error)
	// get status of build identified by given build number
	// outputs one of: success | failed | null
	BuildStatus(ctx context.Context, buildID string) (*string, error)
	// kills running build identified by given build number
	KillBuild(ctx context.Context, buildID string) error
}
