package gitfixture

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitRepository_SubmoduleUpdate(t *testing.T) {
	repo := BasicRepository()
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
}
