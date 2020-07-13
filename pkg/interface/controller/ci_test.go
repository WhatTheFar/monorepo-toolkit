package controller

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelPathsIfPossible(t *testing.T) {
	cases := []*struct {
		workDir  string
		paths    []string
		relPaths []string
	}{
		{
			workDir:  "/Users/far",
			paths:    []string{"/Users/far/README.md", "not/relative"},
			relPaths: []string{"README.md", "not/relative"},
		},
	}

	for i, v := range cases {
		var (
			workDir = v.workDir
			paths   = v.paths
			want    = v.relPaths
		)
		t.Run(fmt.Sprintf("Case %d, ", i), func(t *testing.T) {
			got := relPathsIfPossible(workDir, paths)
			assert.Equal(t, want, got)
		})
	}
}
