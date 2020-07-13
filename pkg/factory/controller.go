package factory

import (
	"os"

	"github.com/whatthefar/monorepo-toolkit/pkg/git"
	interactor_impl "github.com/whatthefar/monorepo-toolkit/pkg/interactor/impl"
	"github.com/whatthefar/monorepo-toolkit/pkg/interface/controller"
	"github.com/whatthefar/monorepo-toolkit/pkg/interface/presenter"
)

func NewCIController(gitWorkDir string, tool string) (controller.CI, error) {
	git, err := git.NewGitGateway(gitWorkDir)
	if err != nil {
		return nil, err
	}
	pipeline, err := NewPipeline(tool)
	if err != nil {
		return nil, err
	}
	presenter := presenter.NewBuildProjectsPresenter(os.Stdout)
	listChangesIt := interactor_impl.NewListChangesInteractor(
		git,
		pipeline,
	)
	buildProjectsIt := interactor_impl.NewBuildProjectsInteractor(
		git,
		pipeline,
		presenter,
	)
	ctrl := controller.NewCIController(listChangesIt, buildProjectsIt)
	return ctrl, nil
}
