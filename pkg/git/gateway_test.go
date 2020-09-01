package git

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
	gitfixture "github.com/whatthefar/monorepo-toolkit/test/git-fixtures"
)

func TestNewGitGateway(t *testing.T) {
	repo := gitfixture.BasicRepository()
	git, err := NewGitGateway(repo.WorkDir())

	assert.NotNil(t, git)
	assert.Nil(t, err)
}

func TestGitGateway_EnsureHavingCommitFromTip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	Convey("Given a newly cloned basic-shallow repository", t, func() {
		ctx := context.Background()
		repo := gitfixture.BasicShallowRepository()

		repo.DeleteWorkDir()
		repo.SubmoduleUpdate()

		git, err := NewGitGateway(repo.WorkDir())

		So(err, ShouldBeNil)

		Convey("When calls EnsureHavingCommitFromTip with a valid SHA", func() {
			sha := core.Hash("64bd0efceae7f8abfd675a2eaadcf3b5aa04e2b1")
			var (
				err error
			)
			err = git.EnsureHavingCommitFromTip(ctx, sha)

			Convey("It should return OK", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When calls EnsureHavingCommitFromTip with an invalid SHA", func() {
			sha := core.Hash("64bd0efceae7f8abfd675")
			var (
				err error
			)
			err = git.EnsureHavingCommitFromTip(ctx, sha)
			isNocommit := git.IsNoCommit(err)

			Convey("It should return error, indicating no commit found", func() {
				So(err, ShouldNotBeNil)
				So(isNocommit, ShouldBeTrue)
			})
		})

		repo.DeleteWorkDir()
		repo.SubmoduleUpdate()
	})
}

func TestGitGateway_hasCommit(t *testing.T) {
	Convey("Given a basic repository", t, func() {
		repo := gitfixture.BasicRepository()
		git, err := NewGitGateway(repo.WorkDir())
		So(err, ShouldBeNil)

		gitImpl, ok := git.(*gitGateway)
		So(ok, ShouldBeTrue)

		Convey("When calls hasCommit func with a valid SHA", func() {
			sha := core.Hash("64bd0efceae7f8abfd675a2eaadcf3b5aa04e2b1")
			var (
				got bool
				err error
			)
			got, err = gitImpl.hasCommit(sha)

			Convey("It should return true", func() {
				So(got, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("When calls hasCommit func with an invalid SHA", func() {
			sha := core.Hash("64bd0efceae7f8abfd675")
			var (
				got bool
				err error
			)
			got, err = gitImpl.hasCommit(sha)

			Convey("It should return true", func() {
				So(got, ShouldBeFalse)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestGitGateway_FilesNameOnly(t *testing.T) {
	Convey("Given a basic repository", t, func() {
		type c struct {
			commit      string
			files       []string
			shouldError bool
		}
		suites := []*struct {
			repo  gitfixture.GitRepository
			cases []*c
		}{
			{
				repo: gitfixture.BasicRepository(),
				cases: []*c{
					{
						commit: "0f998bc84e0b5e764391e22bb554d9705fa7c6c3",
						files: []string{
							"README.md",
							"services/app1/README.md",
							"services/app2/README.md",
							"services/app3/README.md",
						},
					},
					{
						commit: "64bd0efceae7f8abfd675a2eaadcf3b5aa04e2b1",
						files: []string{
							"README.md",
						},
					},
					{
						commit: "55b8c896b86815f519d30c90b431bf8c56bcb278",
						files: []string{
							"README.md",
							"services/app1/README.md",
							"services/app2/README.md",
						},
					},
				},
			},
			{
				repo: gitfixture.BasicShallowRepository(),
				cases: []*c{
					{
						commit: "0f998bc84e0b5e764391e22bb554d9705fa7c6c3",
						files: []string{
							"README.md",
							"services/app1/README.md",
							"services/app2/README.md",
							"services/app3/README.md",
						},
					},
					{
						commit: "64bd0efceae7f8abfd675a2eaadcf3b5aa04e2b1",
						files: []string{
							"README.md",
						},
						shouldError: true,
					},
				},
			},
		}

		for _, s := range suites {
			Convey(fmt.Sprintf("Given a GitGateway for %s repository", s.repo.SubmoduleName()), func() {
				git, err := NewGitGateway(s.repo.WorkDir())

				So(err, ShouldBeNil)

				for i, v := range s.cases {
					var (
						commit = v.commit
						want   = v.files
					)

					Convey(fmt.Sprintf("Case %d, when call FilesNameOnly with commit \"%s\"", i+1, commit), func() {
						got, err := git.FilesNameOnly(core.Hash(commit))

						url := gitfixture.BasicRepository().CommitURL(commit)
						Convey(fmt.Sprintf("Then it should return all changes (%s)", url), func() {
							if v.shouldError == true {
								So(err, ShouldNotBeNil)
								So(got, ShouldBeNil)
							} else {
								So(err, ShouldBeNil)
								So(got, ShouldResemble, want)
							}
						})
					})
				}
			})
		}
	})
}

func TestGitGateway_DiffNameOnly(t *testing.T) {
	Convey("Given a basic repository", t, func() {
		repo := gitfixture.BasicRepository()
		git, err := NewGitGateway(repo.WorkDir())

		So(err, ShouldBeNil)

		cases := []*struct {
			from  string
			to    string
			files []string
		}{
			{
				from: "64bd0efceae7f8abfd675a2eaadcf3b5aa04e2b1", to: "23e1b2860c1a75cbfc6058ca242d5bf25df70c1b",
				files: []string{
					"services/app1/README.md",
					"services/app2/README.md",
					"services/app3/README.md",
				},
			},
			{
				from: "64bd0efceae7f8abfd675a2eaadcf3b5aa04e2b1", to: "eea9c40b4f5093a0bdd4d63c995ef9fc5b76e2b0",
				files: []string{
					"services/app1/README.md",
				},
			},
			{
				from: "eea9c40b4f5093a0bdd4d63c995ef9fc5b76e2b0", to: "55b8c896b86815f519d30c90b431bf8c56bcb278",
				files: []string{
					"services/app2/README.md",
				},
			},
			{
				from: "55b8c896b86815f519d30c90b431bf8c56bcb278", to: "23e1b2860c1a75cbfc6058ca242d5bf25df70c1b",
				files: []string{
					"services/app3/README.md",
				},
			},
			{
				from: "64bd0efceae7f8abfd675a2eaadcf3b5aa04e2b1", to: "0f998bc84e0b5e764391e22bb554d9705fa7c6c3",
				files: []string{
					"services/app1/README.md",
					"services/app2/README.md",
					"services/app3/README.md",
				},
			},
		}

		for i, v := range cases {
			var (
				from = v.from
				to   = v.to
				want = v.files
			)

			Convey(fmt.Sprintf("Case %d, when call DiffNameOnly from \"%s\" to \"%s\"", i+1, from, to), func() {
				got, err := git.DiffNameOnly(core.Hash(from), core.Hash(to))

				url := gitfixture.BasicRepository().CompareURL(from, to)
				Convey(fmt.Sprintf("Then it should return all changes (%s)", url), func() {
					So(err, ShouldBeNil)
					So(got, ShouldResemble, want)
				})
			})
		}
	})
}

func TestGitGateway_DiffNameOnly_Submodules(t *testing.T) {
	Convey("Given a repository with submodules", t, func() {
		repo := gitfixture.SubmodulesRepository()
		git, err := NewGitGateway(repo.WorkDir())

		So(err, ShouldBeNil)

		cases := []*struct {
			from  string
			to    string
			files []string
		}{
			{
				from: "837541c1cc4e7c9895dac58838cafb2de0c64fa9", to: "f5e4c774289e4381e4dde954c774441f61783cc0",
				files: []string{
					".gitmodules",
					"services/app1",
					"services/app2",
				},
			},
			{
				from: "ab010a95d044a8112f0eca9b8b70d375d79a7a37", to: "f5e4c774289e4381e4dde954c774441f61783cc0",
				files: []string{
					"services/app2",
				},
			},
			{
				from: "0f263b81cb8d62d8c43500f74c3319c25744a2bc", to: "ab010a95d044a8112f0eca9b8b70d375d79a7a37",
				files: []string{
					"services/app1",
				},
			},
		}

		for i, v := range cases {
			var (
				from = v.from
				to   = v.to
				want = v.files
			)

			Convey(fmt.Sprintf("Case %d, when call DiffNameOnly from \"%s\" to \"%s\"", i+1, from, to), func() {
				got, err := git.DiffNameOnly(core.Hash(from), core.Hash(to))

				url := gitfixture.BasicRepository().CompareURL(from, to)
				Convey(fmt.Sprintf("Then it should return all changes (%s)", url), func() {
					So(err, ShouldBeNil)
					So(got, ShouldResemble, want)
				})
			})
		}
	})
}
