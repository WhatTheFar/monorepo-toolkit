package pipeline

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
	mock_pipeline "github.com/whatthefar/monorepo-toolkit/pkg/pipeline/mock"
	"github.com/whatthefar/monorepo-toolkit/pkg/utils"
	gitfixture "github.com/whatthefar/monorepo-toolkit/test/git-fixtures"
)

func requireEnv(t *testing.T, key string) string {
	env := os.Getenv(key)
	if env == "" {
		message := fmt.Sprintf("Requires \"%s\" env", key)
		assert.FailNow(t, message)
	}
	return env
}

func TestGetJobIDFromJobURL(t *testing.T) {
	cases := []*struct {
		url         string
		id          int64
		shouldError bool
	}{
		{
			url:         "https://api.github.com/repos/WhatTheFar/monorepo-toolkit/actions/runs/140592597/jobs",
			id:          140592597,
			shouldError: false,
		},
		{
			url:         "https://api.github.com/repos/WhatTheFar/monorepo-toolkit/actions/runs/138239725/jobs",
			id:          138239725,
			shouldError: false,
		},
		{
			url:         "https://api.github.com/repos/WhatTheFar/monorepo toolkit/actions/runs/138239725/jobs",
			id:          0,
			shouldError: true,
		},
		{
			url:         "https://api.github.com/repos/WhatTheFar/monorepo-toolkit/actions/runs/invalid_id/jobs",
			id:          0,
			shouldError: true,
		},
	}

	for i, v := range cases {
		var (
			url         = v.url
			want        = v.id
			shouldError = v.shouldError
		)
		t.Run(fmt.Sprintf("Case %d, calls getJobIDFromJobURL", i+1), func(t *testing.T) {
			id, err := getRunIDFromJobURL(url)
			var got int64 = id

			assert.Equal(t, want, got)
			if shouldError == true {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNewGitHubActionGateway(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	env := mock_pipeline.NewMockGitHubActionEnv(ctrl)

	gw := NewGitHubActionGateway(env)

	assert.Implements(t, (*core.PipelineGateway)(nil), gw)
	assert.IsType(t, new(gitHubActionGateway), gw)

	ghImpl, ok := gw.(*gitHubActionGateway)
	assert.True(t, ok)

	assert.NotNil(t, ghImpl.env)
	assert.Equal(t, env, ghImpl.env)
}

func TestGitHubActionGateway_LastSuccesfulCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	token := requireEnv(t, "GITHUB_TOKEN")

	Convey("Given a GitHubActionGateway", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
		env.EXPECT().Token().Return(token)

		repo := gitfixture.PipelineRepository()

		gw := NewGitHubActionGateway(env)

		cases := []*struct {
			workflowID string
			sha        string
		}{
			{
				workflowID: "build.yml",
				sha:        "ed5434c198b1721d5c83f3f39b6eea967c16f095",
			},
			{
				workflowID: "build-failed.yml",
				sha:        "7163c77dbfb2ed57eab8de7eacc528081eb702c1",
			},
			{
				workflowID: "nop.yml",
				sha:        "",
			},
		}

		for i, v := range cases {
			var (
				workflowID = v.workflowID
				sha        = v.sha
				want       = core.Hash(sha)
			)

			Convey(fmt.Sprintf("Case %d, when LastSuccessfulCommit is called with workflow ID \"%s\", on git-fixture-pipeline", i+1, workflowID), func() {
				env.EXPECT().Owner().Return(repo.Owner())
				env.EXPECT().Repository().Return(repo.Repository())
				env.EXPECT().Branch().Return("master")
				env.EXPECT().Ref().Return("refs/origin/master")
				got, err := gw.LastSuccessfulCommit(ctx, workflowID)

				Convey("Then it should return commit hash with no error", func() {
					So(err, ShouldBeNil)
					So(got, ShouldEqual, want)
				})
			})
		}

	})
}

func TestGitHubActionGateway_CurrentCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	Convey("Given a GitHubActionGateway", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		env := mock_pipeline.NewMockGitHubActionEnv(ctrl)

		gw := NewGitHubActionGateway(env)

		cases := []*struct {
			sha string
		}{
			{sha: "ed5434c198b1721d5c83f3f39b6eea967c16f095"},
			{sha: "7163c77dbfb2ed57eab8de7eacc528081eb702c1"},
		}

		for i, v := range cases {
			var (
				sha  = v.sha
				want = core.Hash(sha)
			)

			Convey(fmt.Sprintf("Case %d, when CurrentCommit is called with sha env \"%s\", on git-fixture-pipeline", i+1, sha), func() {
				env.EXPECT().Sha().Return(sha)
				got := gw.CurrentCommit()

				Convey(fmt.Sprintf("Then it should return commit hash \"%s\" with no error", want), func() {
					So(got, ShouldEqual, want)
				})
			})
		}

	})
}

func TestGitHubActionGateway_TriggerBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	token := requireEnv(t, "GITHUB_TOKEN")

	Convey("Given a GitHubActionGateway", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
		env.EXPECT().Token().Return(token).AnyTimes()

		repo := gitfixture.PipelineRepository()

		gw := NewGitHubActionGateway(env)

		cases := []*struct {
			eventType      string
			projectName    string
			shouldReturnID bool
			shouldError    bool
		}{
			{
				eventType:      "build",
				projectName:    "server",
				shouldReturnID: true,
				shouldError:    false,
			},
			{
				eventType:      "nop",
				projectName:    "server",
				shouldReturnID: false,
				shouldError:    false,
			},
		}

		for i, v := range cases {
			var (
				eventType      = v.eventType
				projectName    = v.projectName
				shouldReturnID = v.shouldReturnID
				shouldError    = v.shouldError
			)

			Convey(fmt.Sprintf(`Case %d, when calls TriggerBuild with project name "%s", on git-fixture-pipeline`, i+1, projectName), func() {
				env.EXPECT().Owner().Return(repo.Owner()).AnyTimes()
				env.EXPECT().Repository().Return(repo.Repository()).AnyTimes()
				env.EXPECT().EventType().Return(eventType)
				var (
					got *string
					err error
				)
				got, err = gw.TriggerBuild(ctx, projectName)

				Convey(fmt.Sprintf("Then it should return build ID"), func() {
					if shouldError == true {
						So(err, ShouldBeError)
					} else {
						So(err, ShouldBeNil)
					}

					if shouldReturnID == true {
						So(got, ShouldNotBeNil)
					} else {
						So(got, ShouldBeNil)
					}
				})
			})
		}

	})
}

func TestGitHubActionGateway_BuildStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	token := requireEnv(t, "GITHUB_TOKEN")

	Convey("Given a GitHubActionGateway", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
		env.EXPECT().Token().Return(token)

		repo := gitfixture.PipelineRepository()

		gw := NewGitHubActionGateway(env)

		cases := []*struct {
			runID  string
			status *string
		}{
			// buld-failed.yml, second run. SHA: 7163c77dbfb2ed57eab8de7eacc528081eb702c1
			{runID: "145647641", status: utils.StrAddr("success")},
			// buld-failed.yml, first run.	SHA: 6e2d4b32f1dae634a08ebe97131276d76e1b11b9
			{runID: "145647981", status: utils.StrAddr("failed")},
		}

		for i, v := range cases {
			var (
				runID = v.runID
				want  = v.status
			)

			Convey(fmt.Sprintf(
				`Case %d, when calls BuildStatus with run ID "%s", on git-fixture-pipeline`,
				i+1,
				runID,
			), func() {
				env.EXPECT().Owner().Return(repo.Owner())
				env.EXPECT().Repository().Return(repo.Repository())
				var (
					got *string
					err error
				)
				got, err = gw.BuildStatus(ctx, runID)

				Convey(fmt.Sprintf("Then it should return status \"%v\"", want), func() {
					So(err, ShouldBeNil)
					So(got, ShouldNotBeNil)
					So(got, ShouldResemble, want)
				})
			})
		}

	})
}

func TestGitHubActionGateway_KillBulid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	token := requireEnv(t, "GITHUB_TOKEN")
	repo := gitfixture.PipelineRepository()

	Convey("Given a GitHubActionGateway", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)

		env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
		env.EXPECT().Token().Return(token).AnyTimes()

		gw := NewGitHubActionGateway(env)

		Convey("Given a build was triggered via TriggerBuild, on git-fixture-pipeline", func() {
			env.EXPECT().Owner().Return(repo.Owner()).MinTimes(1)
			env.EXPECT().Repository().Return(repo.Repository()).MinTimes(1)
			env.EXPECT().EventType().Return("build-kill")

			runID, err := gw.TriggerBuild(ctx, "server")

			So(err, ShouldBeNil)
			So(runID, ShouldNotBeNil)

			ctrl.Finish()

			Convey("Given a new GitHubActionGateway", func() {
				ctx := context.Background()
				ctrl := gomock.NewController(t)

				env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
				env.EXPECT().Token().Return(token)

				gw := NewGitHubActionGateway(env)

				Convey("When calls KillBuild with triggered run ID", func() {
					env.EXPECT().Owner().Return(repo.Owner())
					env.EXPECT().Repository().Return(repo.Repository())

					err := gw.KillBuild(ctx, *runID)

					Convey(fmt.Sprintf("Then it should kill the run"), func() {
						So(err, ShouldBeNil)
						ctrl.Finish()
					})
				})
			})
		})

	})
}
