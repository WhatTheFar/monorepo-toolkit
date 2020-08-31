package gitfixture

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitRepository_IsShallow(t *testing.T) {
	assert.False(t, BasicRepository().IsShallow())
	assert.True(t, BasicShallowRepository().IsShallow())
}

func TestGitRepository_SubmoduleUpdate(t *testing.T) {
	repo := BasicShallowRepository()
	repo.DeleteDotGit()
	repo.DeleteWorkDir()

	var (
		info os.FileInfo
		err  error
	)
	_, err = os.Stat(repo.WorkDir())
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(repo.DotGit())
	assert.True(t, os.IsNotExist(err))

	repo.SubmoduleUpdate()

	info, err = os.Stat(repo.WorkDir())
	assert.NotNil(t, info)
	assert.True(t, info.IsDir())
	assert.False(t, os.IsNotExist(err))

	info, err = os.Stat(repo.DotGit())
	assert.NotNil(t, info)
	assert.True(t, info.IsDir())
	assert.False(t, os.IsNotExist(err))

	// assert if repo still shallow
	assert.True(t, repo.IsShallow())
}
