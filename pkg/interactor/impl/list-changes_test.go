package interactor_impl

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
	. "github.com/whatthefar/monorepo-toolkit/pkg/interactor"
)

const (
	gitFixtureBasicPath = "../../test/git-fixtures/basic"
)

func TestNewListChnagesInteractor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	git := mock_core.NewMockGitGateway(ctrl)
	pipeline := mock_core.NewMockPipelineGateway(ctrl)
	interactor := NewListChangesInteractor(git, pipeline)

	assert.Implements(t, (*ListChangesInteractor)(nil), interactor)
	assert.IsType(t, new(listChangesInteractor), interactor)

	impl, ok := interactor.(*listChangesInteractor)
	assert.True(t, ok)

	assert.NotNil(t, impl.git)
	assert.Equal(t, git, impl.git)
	assert.NotNil(t, impl.pipeline)
	assert.Equal(t, pipeline, impl.pipeline)
}

func TestListChangesInteractor(t *testing.T) {
	Convey("Given a listChangesInteractor", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		git := mock_core.NewMockGitGateway(ctrl)
		pipeline := mock_core.NewMockPipelineGateway(ctrl)

		interactor := &listChangesInteractor{git, pipeline}

		workflowID := "main.yml"
		var lastCommit core.Hash
		var currentCommit core.Hash

		ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()

		cases := []*struct {
			paths         []string
			lastCommit    string
			currentCommit string
			want          []string
			expect        func()
		}{
			{
				paths:         []string{"services/app1"},
				lastCommit:    "123",
				currentCommit: "456",
				want:          []string{"services/app1"},
				expect: func() {
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
				},
			},
			{
				paths: []string{"services/app1"},
				// no lastSuccessfulCommit
				lastCommit:    "",
				currentCommit: "456",
				want:          []string{"services/app1"},
				expect: func() {
					pipeline.EXPECT().
						LastSuccessfulCommit(gomock.AssignableToTypeOf(ctxType), gomock.Eq(workflowID)).
						Return(lastCommit, nil)
					pipeline.EXPECT().
						CurrentCommit().
						Return(currentCommit)

					git.EXPECT().
						FilesNameOnly(
							gomock.Eq(currentCommit),
						).
						Return([]string{
							"services/app1/README.md",
							"services/app2/README.md",
							"services/app3/README.md",
						}, nil)
				},
			},
		}

		for i, v := range cases {
			var (
				paths  = v.paths
				want   = v.want
				expect = v.expect
			)
			lastCommit = core.Hash(v.lastCommit)
			currentCommit = core.Hash(v.currentCommit)

			Convey(fmt.Sprintf("Case %d, given mocks", i), func() {
				expect()

				Convey("When calls Execute", func() {
					got, err := interactor.ListChanges(ctx, paths, workflowID)

					So(err, ShouldBeNil)
					So(got, ShouldResemble, want)
				})
			})
		}
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
