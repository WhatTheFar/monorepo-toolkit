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
	cases := []struct {
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
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
	env.EXPECT().Token().Return("github_personal_access_token")

	gw := NewGitHubActionGateway(ctx, env)

	assert.Implements(t, (*core.PipelineGateway)(nil), gw)
	assert.IsType(t, new(gitHubActionGateway), gw)

	ghImpl, ok := gw.(*gitHubActionGateway)
	assert.True(t, ok)

	assert.NotNil(t, ghImpl.client)
	assert.NotNil(t, ghImpl.env)
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

		gw := NewGitHubActionGateway(ctx, env)

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

	token := requireEnv(t, "GITHUB_TOKEN")

	Convey("Given a GitHubActionGateway", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
		env.EXPECT().Token().Return(token)

		gw := NewGitHubActionGateway(ctx, env)

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

			Convey(fmt.Sprintf("Case %d, when LastSuccessfulCommit is called with sha env \"%s\", on git-fixture-pipeline", i+1, sha), func() {
				env.EXPECT().Sha().Return(sha)
				got := gw.CurrentCommit()

				Convey(fmt.Sprintf("Then it should return commit hash \"%s\" with no error", want), func() {
					So(got, ShouldEqual, want)
				})
			})
		}

	})
}
