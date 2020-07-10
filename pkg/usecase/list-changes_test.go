package usecase

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
	mock_core "github.com/whatthefar/monorepo-toolkit/pkg/core/mock"
)

const (
	gitFixtureBasicPath = "../../test/git-fixtures/basic"
)

func TestListChangesUseCase(t *testing.T) {
	Convey("Given a listChangesUseCase", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		git := mock_core.NewMockGitGateway(ctrl)
		pipeline := mock_core.NewMockPipelineGateway(ctrl)

		uc := &listChangesUseCase{git, pipeline}

		Convey("When calls Execute", func() {
			paths := []string{"services/app1"}
			workflowID := "main.yml"

			lastCommit := core.Hash("123")
			currentCommit := core.Hash("456")

			ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
			pipeline.EXPECT().
				LastSuccessfulCommit(gomock.AssignableToTypeOf(ctxType), gomock.Eq(workflowID)).
				Return(lastCommit, nil)
			pipeline.EXPECT().
				CurrentCommit().
				Return(currentCommit)
			git.EXPECT().
				EnsureHavingCommitFromTip(
					gomock.AssignableToTypeOf(ctxType),
					gomock.Eq(lastCommit),
				).
				Return(nil)
			git.EXPECT().
				DiffNameOnly(
					gomock.Eq(lastCommit),
					gomock.Eq(currentCommit),
				).
				Return([]string{
					"services/app1/README.md",
					"services/app2/README.md",
					"services/app3/README.md",
				}, nil)

			want := []string{
				"services/app1",
			}

			got, err := uc.ListChanges(ctx, paths, workflowID)

			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}

func TestFilterOnlyPathsWithChanges(t *testing.T) {
	cases := []*struct {
		paths    []string
		changes  []string
		expected []string
	}{
		{
			paths: []string{"services/app1", "services/app2"},
			changes: []string{
				"services/app1/README.md",
				"services/app2/README.md",
				"services/app3/README.md",
			},
			expected: []string{"services/app1", "services/app2"},
		},
		{
			paths: []string{"services/app1/README.md", "services/app2/README.md"},
			changes: []string{
				"services/app1/README.md",
				"services/app2/README.md",
				"services/app3/README.md",
			},
			expected: []string{"services/app1/README.md", "services/app2/README.md"},
		},
		{
			paths: []string{"pkg", "services/app3"},
			changes: []string{
				"services/app1/README.md",
				"services/app2/README.md",
				"services/app3/README.md",
			},
			expected: []string{"services/app3"},
		},
	}

	for i, v := range cases {
		var (
			paths   = v.paths
			changes = v.changes
			want    = v.expected
		)
		t.Run(fmt.Sprintf("Case %d, filterOnlyPathsWithChanges should work", i), func(t *testing.T) {
			var got []string
			got = filterOnlyPathsWithChanges(paths, changes)

			assert.Equal(t, want, got)
		})
	}
}
