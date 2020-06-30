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
		cases := []map[string]string{
			{
				GITHUB_TOKEN:      "1234567890",
				GITHUB_REF:        "refs/heads/master",
				GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
				GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
				GITHUB_EVENT_TYPE: "build",
			},
			{
				GITHUB_TOKEN:      "1234567890",
				GITHUB_REF:        "refs/heads/master",
				GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
				GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
				// empty event type
				GITHUB_EVENT_TYPE: "",
			},
			{
				GITHUB_TOKEN:      "1234567890",
				GITHUB_REF:        "refs/heads/master",
				GITHUB_SHA:        "0770df1c082d9e0e3aaf1a32ad65d8b5006964f6",
				GITHUB_REPOSITORY: "WhatTheFar/monorepo-toolkit",
				// no event type
			},
		}

		for i, env := range cases {
			Convey(fmt.Sprintf("Case %d, setup env variables", i), func() {
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

				Convey("Assert NewGitHubActionEnv", func() {
					var ghEnv GitHubActionEnv = NewGitHubActionEnv()

					assert.NotNil(t, ghEnv)
					assert.Implements(t, (*GitHubActionEnv)(nil), ghEnv)
					assert.IsType(t, new(gitHubActionEnv), ghEnv)

					assert.Equal(t, env[GITHUB_TOKEN], ghEnv.Token())
					assert.Equal(t, env[GITHUB_REF], ghEnv.Ref())
					assert.Equal(t, "master", ghEnv.Branch())
					assert.Equal(t, env[GITHUB_SHA], ghEnv.Sha())
					assert.Equal(t, "WhatTheFar", ghEnv.Owner())
					assert.Equal(t, "monorepo-toolkit", ghEnv.Repository())
					assert.Equal(t, env[GITHUB_EVENT_TYPE], ghEnv.EventType())
				})
			})
		}
	})

}
