package factory

import (
	"errors"
	"fmt"

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
		return pipeline.NewGitHubActionGateway(env), nil
	case "travis":
		return nil, errors.New(fmt.Sprintf(`CI_TOOl "%s" is not currently supported`, tool))
	default:
		return nil, errors.New(fmt.Sprintf(`CI_TOOL "%s" is invalid or not unsupported`, tool))
	}
}
