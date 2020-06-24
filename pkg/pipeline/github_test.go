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
)

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

	token := os.Getenv("GITHUB_TOKEN")
	assert.NotEmpty(t, token, "Requires GITHUB_TOKEN env")

	Convey("Given a GitHubActionGateway", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		env := mock_pipeline.NewMockGitHubActionEnv(ctrl)
		env.EXPECT().Token().Return(token)

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
				env.EXPECT().Owner().Return("WhatTheFar")
				env.EXPECT().Repository().Return("monorepo-toolkit-git-fixture-pipeline")
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
