package pipeline

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewGitHubActionEnv(t *testing.T) {
	Convey("Setup", t, func() {
		cases := []*struct {
			env        map[string]string
			branch     string
			owner      string
			repository string
			isValid    bool
		}{
			{
				env: map[string]string{
					GITHUB_TOKEN:      "1234567890",
					GITHUB_REF:        "refs/heads/master",
					GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
					GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
					GITHUB_EVENT_TYPE: "build",
				},
				branch:     "master",
				owner:      "WhatTheFar",
				repository: "monorepo-toolkit",
				isValid:    true,
			},
			{
				env: map[string]string{
					GITHUB_TOKEN:      "1234567890",
					GITHUB_REF:        "refs/heads/master",
					GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
					GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
					// empty event type
					GITHUB_EVENT_TYPE: "",
				},
				branch:     "master",
				owner:      "WhatTheFar",
				repository: "monorepo-toolkit",
				isValid:    true,
			},
			{
				env: map[string]string{
					GITHUB_TOKEN:      "1234567890",
					GITHUB_REF:        "refs/heads/master",
					GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
					GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
					// no event type
				},
				branch:     "master",
				owner:      "WhatTheFar",
				repository: "monorepo-toolkit",
				isValid:    true,
			},
			{
				// try another combination of env values
				env: map[string]string{
					GITHUB_TOKEN: "0987654321",
					GITHUB_REF:   "refs/heads/develop",
					// missing GITHUB_SHA
					GITHUB_REPOSITORY: "whatthefar/monorepo-toolkit-git-fixture-basic",
					// no event type
				},
				branch:     "develop",
				owner:      "whatthefar",
				repository: "monorepo-toolkit-git-fixture-basic",
				isValid:    false,
			},
			{
				env: map[string]string{
					// missing GITHUB_TOKEN
					GITHUB_REF:        "refs/heads/master",
					GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
					GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
					// no event type
				},
				branch:     "master",
				owner:      "WhatTheFar",
				repository: "monorepo-toolkit",
				isValid:    false,
			},
			{
				env: map[string]string{
					GITHUB_TOKEN: "1234567890",
					// missing GITHUB_REF
					GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
					GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
					// no event type
				},
				branch:     "",
				owner:      "WhatTheFar",
				repository: "monorepo-toolkit",
				isValid:    false,
			},
			{
				env: map[string]string{
					GITHUB_TOKEN: "1234567890",
					GITHUB_REF:   "refs/heads/master",
					GITHUB_SHA:   "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
					// missing GITHUB_REPOSITORY
					// no event type
				},
				branch:     "master",
				owner:      "",
				repository: "",
				isValid:    false,
			},
		}

		for i, v := range cases {
			var (
				env        = v.env
				branch     = v.branch
				owner      = v.owner
				repository = v.repository
				isValid    = v.isValid
			)
			Convey(fmt.Sprintf("Case %d, given env variables", i), func() {
				envBackup := make(map[string]string)
				defer func() {
					// restore env varaibles
					for k, v := range envBackup {
						os.Setenv(k, v)
					}
				}()
				for k, v := range env {
					// backup env variables
					envBackup[k] = os.Getenv(k)
					os.Setenv(k, v)
				}

				Convey("When calls NewGitHubActionEnv", func() {
					var ghEnv GitHubActionEnv = NewGitHubActionEnv()

					Convey("It should return a GitHubActionEnv with correct values", func() {
						assert.NotNil(t, ghEnv)
						assert.Implements(t, (*GitHubActionEnv)(nil), ghEnv)
						assert.IsType(t, new(gitHubActionEnv), ghEnv)

						assert.Equal(t, env[GITHUB_TOKEN], ghEnv.Token())
						assert.Equal(t, env[GITHUB_REF], ghEnv.Ref())
						assert.Equal(t, branch, ghEnv.Branch())
						assert.Equal(t, env[GITHUB_SHA], ghEnv.Sha())
						assert.Equal(t, owner, ghEnv.Owner())
						assert.Equal(t, repository, ghEnv.Repository())
						assert.Equal(t, env[GITHUB_EVENT_TYPE], ghEnv.EventType())

					})

					Convey("And calls env.Validate()", func() {
						err := ghEnv.Validate()

						Convey("It should return correct response", func() {
							if isValid == true {
								So(err, ShouldBeNil)
							} else {
								So(err, ShouldBeError)
							}
						})
					})
				})
			})
		}
	})

}
