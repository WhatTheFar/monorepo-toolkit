package pipeline

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	GITHUB_TOKEN      = "GITHUB_TOKEN"
	GITHUB_REF        = "GITHUB_REF"
	GITHUB_SHA        = "GITHUB_SHA"
	GITHUB_REPOSITORY = "GITHUB_REPOSITORY"
	GITHUB_EVENT_TYPE = "GITHUB_EVENT_TYPE"

	gitHubRefSeparator        = "/"
	gitHubRepositorySeparator = "/"
)

func NewGitHubActionEnv() GitHubActionEnv {
	env := &gitHubActionEnv{}
	v := viper.New()
	v.BindEnv(GITHUB_TOKEN)
	v.BindEnv(GITHUB_REF)
	v.BindEnv(GITHUB_SHA)
	v.BindEnv(GITHUB_REPOSITORY)
	v.BindEnv(GITHUB_EVENT_TYPE)
	v.AllowEmptyEnv(true)
	v.Unmarshal(env)
	return env
}

type gitHubActionEnv struct {
	GitHubToken      string `mapstructure:"GITHUB_TOKEN"`
	GitHubRef        string `mapstructure:"GITHUB_REF"`
	GitHubSha        string `mapstructure:"GITHUB_SHA"`
	GitHubRepository string `mapstructure:"GITHUB_REPOSITORY"`
	GitHubEventType  string `mapstructure:"GITHUB_EVENT_TYPE"`
}

func (e *gitHubActionEnv) Token() string {
	return e.GitHubToken
}

func (e *gitHubActionEnv) Ref() string {
	return e.GitHubRef
}

func (e *gitHubActionEnv) Branch() string {
	s := strings.Split(e.GitHubRef, gitHubRefSeparator)
	return s[len(s)-1]
}

func (e *gitHubActionEnv) Sha() string {
	return e.GitHubSha
}

func (e *gitHubActionEnv) Owner() string {
	s := strings.Split(e.GitHubRepository, gitHubRepositorySeparator)
	return s[len(s)-2]
}

func (e *gitHubActionEnv) Repository() string {
	s := strings.Split(e.GitHubRepository, gitHubRepositorySeparator)
	return s[len(s)-1]
}

func (e *gitHubActionEnv) EventType() string {
	return e.GitHubEventType
}
