package factory

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/whatthefar/monorepo-toolkit/pkg/core"
	"github.com/whatthefar/monorepo-toolkit/pkg/pipeline"
)

func NewPipeline(tool string) (core.PipelineGateway, error) {
	switch tool {
	case "bitbucket":
		return nil, errors.New(fmt.Sprintf(`CI_TOOl "%s" is not currently supported`, tool))
	case "circleci":
		return nil, errors.New(fmt.Sprintf(`CI_TOOl "%s" is not currently supported`, tool))
	case "github":
		env := pipeline.NewGitHubActionEnv()
		err := env.Validate()
		if err != nil {
			return nil, errors.Wrap(err, "fail to validate envs for github action")
		}
		return pipeline.NewGitHubActionGateway(env), nil
	case "travis":
		return nil, errors.New(fmt.Sprintf(`CI_TOOl "%s" is not currently supported`, tool))
	default:
		return nil, errors.New(fmt.Sprintf(`CI_TOOL "%s" is invalid or not unsupported`, tool))
	}
}
