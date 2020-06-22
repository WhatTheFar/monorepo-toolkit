package git

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/whatthefar/monorepo-toolkit/pkg/core"
)

var (
	GIT_FIXTURE_BASIC_REPOSITORY = "https://github.com/WhatTheFar/monorepo-toolkit-git-fixture-basic"

	GIT_FIXTURE_BASIC_PATH = "../../test/git-fixtures/basic"
)

func compareUrl(repo, from, to string) string {
	return fmt.Sprintf("%s/compare/%s..%s", repo, from, to)
}

func TestNewGitGateway(t *testing.T) {
	git, err := NewGitGateway(GIT_FIXTURE_BASIC_PATH)

	assert.NotNil(t, git)
	assert.Nil(t, err)
}

func TestGitGateway(t *testing.T) {
	Convey("Given a basic repository", t, func() {
		git, err := NewGitGateway(GIT_FIXTURE_BASIC_PATH)

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

				url := compareUrl(GIT_FIXTURE_BASIC_REPOSITORY, from, to)
				Convey(fmt.Sprintf("Then it should return all changes (%s)", url), func() {
					So(err, ShouldBeNil)
					So(got, ShouldResemble, want)
				})
			})
		}
	})
}
