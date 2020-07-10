package usecase

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	mock_core "github.com/whatthefar/monorepo-toolkit/pkg/core/mock"
	mock_interactor "github.com/whatthefar/monorepo-toolkit/pkg/interactor/mock"
	"github.com/whatthefar/monorepo-toolkit/pkg/utils"
)

func TestNewBuildProjectsInteractor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	git := mock_core.NewMockGitGateway(ctrl)
	pipeline := mock_core.NewMockPipelineGateway(ctrl)
	presenter := mock_interactor.NewMockBuildProjectsPresenter(ctrl)
	interactor := NewBuildProjectsInteractor(git, pipeline, presenter)

	assert.Implements(t, (*BuildProjectsInteractor)(nil), interactor)
	assert.IsType(t, new(buildProjectsInteractor), interactor)

	impl, ok := interactor.(*buildProjectsInteractor)
	assert.True(t, ok)

	assert.NotNil(t, impl.ListChangesInteractor)
	assert.NotNil(t, impl.iListProjects)
	assert.NotNil(t, impl.pipeline)
	assert.NotNil(t, impl.presenter)
}

func TestNewBuildProjectsOnceInteractor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	git := mock_core.NewMockGitGateway(ctrl)
	pipeline := mock_core.NewMockPipelineGateway(ctrl)
	presenter := mock_interactor.NewMockBuildProjectsPresenter(ctrl)
	interactor := NewBuildProjectsOnceInteractor(git, pipeline, presenter)

	assert.Implements(t, (*BuildProjectsInteractor)(nil), interactor)
	assert.IsType(t, new(buildProjectsInteractor), interactor)

	impl, ok := interactor.(*buildProjectsInteractor)
	assert.True(t, ok)

	assert.NotNil(t, impl.ListChangesInteractor)
	assert.NotNil(t, impl.iListProjects)
	assert.NotNil(t, impl.pipeline)
	assert.NotNil(t, impl.presenter)
}

func TestBuildProjectsInteractor(t *testing.T) {
	Convey("Given a buildProjectsInteractor", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		// defer ctrl.Finish()

		pipeline := mock_core.NewMockPipelineGateway(ctrl)

		listChangesUc := mock_interactor.NewMockListChangesInteractor(ctrl)
		presenter := mock_interactor.NewMockBuildProjectsPresenter(ctrl)
		interactor := &buildProjectsInteractor{listChangesUc, &listProjects{}, presenter, pipeline}

		// reset constant to default value
		buildMaxSeconds = buildMaxSecondsDefault
		buildCheckAfterSeconds = buildCheckAfterSecondsDefault

		Convey("Mock a ListChanges func", func() {
			paths := []string{"services/app1", "services/app2"}
			projectNames := []string{"app1", "app2"}
			buildIDs := []string{"111", "222"}
			workflowID := "main.yml"
			ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()

			listChangesUc.EXPECT().
				ListChanges(
					gomock.AssignableToTypeOf(ctxType),
					gomock.Eq(paths),
					gomock.Eq(workflowID),
				).
				Return(paths, nil)

			Convey("Setup successful builds", func() {
				for i, name := range projectNames {
					buildID := buildIDs[i]
					pipeline.EXPECT().
						TriggerBuild(
							gomock.AssignableToTypeOf(ctxType),
							gomock.Eq(name),
						).
						Return(utils.StrAddr(buildID), nil)
					presenter.EXPECT().BuildTriggeredFor(name).Return()

					pipeline.EXPECT().
						BuildStatus(
							gomock.AssignableToTypeOf(ctxType),
							gomock.Eq(buildID),
						).
						Return(utils.StrAddr("success"), nil)
				}
				presenter.EXPECT().AllBuildSucceeded()

				Convey("When BuildFor is called", func() {
					interactor.BuildFor(ctx, paths, workflowID)

					Convey("All expectaton should pass", func() {
						ctrl.Finish()
					})
				})
			})

			Convey("Setup second build to fail", func() {
				for i, name := range projectNames {
					buildID := buildIDs[i]
					pipeline.EXPECT().
						TriggerBuild(
							gomock.AssignableToTypeOf(ctxType),
							gomock.Eq(name),
						).
						Return(utils.StrAddr(buildID), nil)
					presenter.EXPECT().BuildTriggeredFor(name).Return()

					if i == 1 {
						// fail build status
						pipeline.EXPECT().
							BuildStatus(
								gomock.AssignableToTypeOf(ctxType),
								gomock.Eq(buildID),
							).
							Return(utils.StrAddr("failed"), nil)
						presenter.EXPECT().BuildFailedFor(name)
					} else {
						// success build status
						pipeline.EXPECT().
							BuildStatus(
								gomock.AssignableToTypeOf(ctxType),
								gomock.Eq(buildID),
							).
							Return(utils.StrAddr("success"), nil)
					}
				}

				Convey("When BuildFor is called", func() {
					interactor.BuildFor(ctx, paths, workflowID)

					Convey("All expectaton should pass", func() {
						ctrl.Finish()
					})
				})
			})

			Convey("Setup no build triggered", func() {
				for _, name := range projectNames {
					// no build ID, on TriggerBuild
					pipeline.EXPECT().
						TriggerBuild(
							gomock.AssignableToTypeOf(ctxType),
							gomock.Eq(name),
						).
						Return(nil, nil)
					presenter.EXPECT().NoBuildTriggeredFor(name).Return()
				}
				presenter.EXPECT().AllBuildSucceeded().Return()

				Convey("When BuildFor is called", func() {
					interactor.BuildFor(ctx, paths, workflowID)

					Convey("All expectaton should pass", func() {
						ctrl.Finish()
					})
				})
			})

			Convey("Setup wait for second build until timeout", func() {
				buildMaxSeconds = 5*time.Second + 500*time.Millisecond
				buildCheckAfterSeconds = 1 * time.Second

				for i, name := range projectNames {
					buildID := buildIDs[i]
					pipeline.EXPECT().
						TriggerBuild(
							gomock.AssignableToTypeOf(ctxType),
							gomock.Eq(name),
						).
						Return(utils.StrAddr(buildID), nil)
					presenter.EXPECT().BuildTriggeredFor(name).Return()

					if i == 1 {
						// no build status
						pipeline.EXPECT().
							BuildStatus(
								gomock.AssignableToTypeOf(ctxType),
								gomock.Eq(buildID),
							).
							Return(nil, nil).
							MinTimes(6).MaxTimes(6)
					} else {
						// success build status
						pipeline.EXPECT().
							BuildStatus(
								gomock.AssignableToTypeOf(ctxType),
								gomock.Eq(buildID),
							).
							Return(utils.StrAddr("success"), nil)
					}
				}
				// waiting for both project
				presenter.EXPECT().WaitingFor(projectNames[1:]).Return().
					MinTimes(6).MaxTimes(6)
				// should timeout
				presenter.EXPECT().Timeout().Return()

				// it should kill all running builds
				presenter.EXPECT().KillingBuildsFor(projectNames[1:]).Return()
				pipeline.EXPECT().
					KillBuild(
						gomock.AssignableToTypeOf(ctxType),
						gomock.Eq(buildIDs[1]),
					).
					Return(nil)
				// all running builds should have been killed
				presenter.EXPECT().NotFinishedBuildsKilled().Return()

				Convey("When BuildFor is called", func() {
					interactor.BuildFor(ctx, paths, workflowID)

					Convey("All expectaton should pass", func() {
						ctrl.Finish()
					})
				})
			})
		})
	})
}
