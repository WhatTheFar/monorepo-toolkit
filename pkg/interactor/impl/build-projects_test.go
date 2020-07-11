package interactor_impl

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	mock_core "github.com/whatthefar/monorepo-toolkit/pkg/core/mock"
	. "github.com/whatthefar/monorepo-toolkit/pkg/interactor"
	mock_interactor "github.com/whatthefar/monorepo-toolkit/pkg/interactor/mock"
	"github.com/whatthefar/monorepo-toolkit/pkg/utils"
)

func TestNewBuildProjectsInteractor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	git := mock_core.NewMockGitGateway(ctrl)
	pipeline := mock_core.NewMockPipelineGateway(ctrl)
	presenter := mock_interactor.NewMockBuildProjectsOutput(ctrl)
	interactor := NewBuildProjectsInteractor(git, pipeline, presenter)

	assert.Implements(t, (*BuildProjectsInteractor)(nil), interactor)
	assert.IsType(t, new(buildProjectsInteractor), interactor)

	impl, ok := interactor.(*buildProjectsInteractor)
	assert.True(t, ok)

	assert.NotNil(t, impl.ListChangesInteractor)
	assert.NotNil(t, impl.presenter)
	assert.Equal(t, presenter, impl.presenter)
	assert.NotNil(t, impl.pipeline)
	assert.Equal(t, pipeline, impl.pipeline)
}

func TestBuildProjectsInteractor(t *testing.T) {
	Convey("Given a buildProjectsInteractor", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)

		pipeline := mock_core.NewMockPipelineGateway(ctrl)

		listChangesUc := mock_interactor.NewMockListChangesInteractor(ctrl)
		presenter := mock_interactor.NewMockBuildProjectsOutput(ctrl)
		interactor := &buildProjectsInteractor{
			ListChangesInteractor: listChangesUc,
			presenter:             presenter,
			pipeline:              pipeline,
		}

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
					presenter.EXPECT().BuildTriggeredFor(name, buildID).Return()

					pipeline.EXPECT().
						BuildStatus(
							gomock.AssignableToTypeOf(ctxType),
							gomock.Eq(buildID),
						).
						Return(utils.StrAddr("success"), nil)
				}
				presenter.EXPECT().AllBuildSucceeded(projectNames)

				Convey("When BuildFor is called", func() {
					interactor.BuildPaths(ctx, paths, workflowID)

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
					presenter.EXPECT().BuildTriggeredFor(name, buildID).Return()

					if i == 1 {
						// fail build status
						pipeline.EXPECT().
							BuildStatus(
								gomock.AssignableToTypeOf(ctxType),
								gomock.Eq(buildID),
							).
							Return(utils.StrAddr("failed"), nil)
						presenter.EXPECT().BuildFailedFor(name, buildID)
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
					interactor.BuildPaths(ctx, paths, workflowID)

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
				presenter.EXPECT().AllBuildSucceeded(projectNames).Return()

				Convey("When BuildFor is called", func() {
					interactor.BuildPaths(ctx, paths, workflowID)

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
					presenter.EXPECT().BuildTriggeredFor(name, buildID).Return()

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
				infos := []*BuildInfo{{ProjectName: projectNames[1], BuildID: buildIDs[1]}}
				presenter.EXPECT().WaitingFor(infos).Return().
					MinTimes(6).MaxTimes(6)
				// should timeout
				presenter.EXPECT().Timeout(buildMaxSeconds).Return()

				// it should kill all running builds
				presenter.EXPECT().KillingBuilds(infos).Return()
				pipeline.EXPECT().
					KillBuild(
						gomock.AssignableToTypeOf(ctxType),
						gomock.Eq(buildIDs[1]),
					).
					Return(nil)
				// all running builds should have been killed
				presenter.EXPECT().NotFinishedBuildsKilled().Return()

				Convey("When BuildFor is called", func() {
					interactor.BuildPaths(ctx, paths, workflowID)

					Convey("All expectaton should pass", func() {
						ctrl.Finish()
					})
				})
			})
		})
	})
}
