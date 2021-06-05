package cmd

import (
	"fmt"
	"os"

	"github.com/whatthefar/monorepo-toolkit/pkg/factory"
)

var (
	ciControllerFactory = factory.CIController
)

func er(msg interface{}) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}
