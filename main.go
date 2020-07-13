package main

import (
	"os"

	"github.com/whatthefar/monorepo-toolkit/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		// fmt.Println("Error:", err)
		os.Exit(1)
	}
}
