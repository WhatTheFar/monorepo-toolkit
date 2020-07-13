package cmd

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cobra"
	factory_mock "github.com/whatthefar/monorepo-toolkit/pkg/factory/mock"
	mock_controller "github.com/whatthefar/monorepo-toolkit/pkg/interface/controller/mock"
)

func TestBuildCmdFlag(t *testing.T) {
	Convey("Given a monorepo-toolkit command", t, func() {
		cmd := newMonorepoToolkit()
		Convey("Given a run func is mocked", func() {
			var flags *buildCmdFlag
			buildCmd.Run = func(cmd *cobra.Command, args []string) {
				f := newBuildCmdFlag()
				flags = f
			}

			cases := []*struct {
				args       []string
				tool       string
				workflowID string
				once       bool
			}{
				{
					args: []string{
						"build",
						"--ci-tool", "github",
						"--workflow", "main.yml",
						"--once",
						"services",
					},
					tool:       "github",
					workflowID: "main.yml",
					once:       true,
				},
				{
					// no once flags
					args: []string{
						"build",
						"--ci-tool", "github",
						"--workflow", "main.yml",
						"services",
					},
					tool:       "github",
					workflowID: "main.yml",
					once:       false,
				},
			}

			for i, v := range cases {
				var (
					args       = v.args
					tool       = v.tool
					workflowID = v.workflowID
					once       = v.once
				)

				Convey(fmt.Sprintf("Case %d, when execute cmd with flags", i), func() {
					cmd.SetArgs(args)
					err := cmd.Execute()

					Convey("It should unmarhsal flags", func() {
						So(err, ShouldBeNil)
						So(flags.CITool, ShouldEqual, tool)
						So(flags.WorkflowID, ShouldEqual, workflowID)
						So(flags.Once, ShouldEqual, once)
					})
				})
			}
		})
	})
}

func TestBuildProjectsCmd(t *testing.T) {
	Convey("Given a monorepo-toolkit command", t, func() {
		cmd := newMonorepoToolkit()
		Convey("Given a mock CI controller factory", func() {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			ciContoller := mock_controller.NewMockCI(ctrl)
			factory := factory_mock.NewMockCIControllerFactory(ctrl)
			// replace default factory with the mock one
			ciControllerFactory = factory

			cases := []*struct {
				args   []string
				tool   string
				expect func()
			}{
				{
					args: []string{
						"build",
						"--ci-tool", "github",
						"--workflow", "main.yml",
						"--once",
						"services",
					},
					tool: "github",
					expect: func() {
						ciContoller.EXPECT().
							BuildOnce(ctx, []string{"services"}, "main.yml")
					},
				},
				{
					// no once flags
					args: []string{
						"build",
						"--ci-tool", "github",
						"--workflow", "main.yml",
						"services",
					},
					tool: "github",
					expect: func() {
						ciContoller.EXPECT().
							Build(ctx, []string{"services"}, "main.yml")
					},
				},
			}

			for i, v := range cases {
				var (
					args   = v.args
					tool   = v.tool
					expect = v.expect
				)

				Convey(fmt.Sprintf("Case %d, given ci controller and factory are mocked", i), func() {
					wd, err := os.Getwd()
					So(err, ShouldBeNil)
					factory.EXPECT().New(wd, tool).Return(ciContoller, nil)
					expect()

					Convey("When execute cmd with args", func() {

						cmd.SetArgs(args)
						err = cmd.Execute()

						Convey("All expectations should pass", func() {
							So(err, ShouldBeNil)
							ctrl.Finish()
						})
					})
				})
			}

		})
	})
}
