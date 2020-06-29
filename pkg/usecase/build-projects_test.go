package usecase

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	mock_core "github.com/whatthefar/monorepo-toolkit/pkg/core/mock"
	mock_usecase "github.com/whatthefar/monorepo-toolkit/pkg/usecase/mock"
	"github.com/whatthefar/monorepo-toolkit/pkg/utils"
)

func TestBuildProjectsUseCase(t *testing.T) {
	Convey("Given a buildProjectsUseCase", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		// defer ctrl.Finish()

		pipeline := mock_core.NewMockPipelineGateway(ctrl)

		listChangesUc := mock_usecase.NewMockListChangesUseCase(ctrl)
		presenter := mock_usecase.NewMockBuildProjectsPresenter(ctrl)
		uc := &buildProjectsUseCase{listChangesUc, presenter, pipeline}

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
					uc.BuildFor(ctx, paths, workflowID)

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
					uc.BuildFor(ctx, paths, workflowID)

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
					uc.BuildFor(ctx, paths, workflowID)

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
						for i := 0; i < 6; i++ {
							pipeline.EXPECT().
								BuildStatus(
									gomock.AssignableToTypeOf(ctxType),
									gomock.Eq(buildID),
								).
								Return(nil, nil)
						}
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
				for i := 0; i < 6; i++ {
					// waiting for both project
					presenter.EXPECT().WaitingFor(projectNames[1:]).Return()
				}
				presenter.EXPECT().Timeout().Return()

				Convey("When BuildFor is called", func() {
					uc.BuildFor(ctx, paths, workflowID)

					Convey("All expectaton should pass", func() {
						ctrl.Finish()
					})
				})
			})
		})
	})
}
