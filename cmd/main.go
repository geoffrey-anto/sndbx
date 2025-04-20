package main

import (
	"fmt"
	"os"
	"path/filepath"

	check_current_context "github.com/geoffrey-anto/sndbx/internal/check_context"
	"github.com/geoffrey-anto/sndbx/internal/sandbox"
)

func main() {
	exists, file := check_current_context.CheckIfDockerfileExists()

	if !exists {
		fmt.Printf("Available images %+v\n", sandbox.GetAvailableEnvironments())
	} else {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		sandbox.CreateContainerWithLocalDockerfile(file, filepath.Base(dir))
	}
}
