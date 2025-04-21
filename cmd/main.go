package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	check_current_context "github.com/geoffrey-anto/sndbx/internal/check_context"
	"github.com/geoffrey-anto/sndbx/internal/sandbox"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "sndbx",
		Usage: "Spawn a quick sandbox 📦✅",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialize something",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "build",
						DefaultText: "false",
						Aliases:     []string{"b"},
						Usage:       "Enable build mode",
						Value:       false,
					},
					&cli.StringFlag{
						Name:    "context",
						Aliases: []string{"c"},
						Usage:   "Name of Dockerfile/Image",
						Value:   "",
					},
					&cli.BoolFlag{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "Remove the sandbox after use",
						Value:   false,
					},
					&cli.IntSliceFlag{
						Name:    "ports",
						Aliases: []string{"p"},
						Usage:   "Ports to expose",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("build") {
						if c.String("context") == "" {
							exists, dockerfile := check_current_context.CheckIfDockerfileExists()

							if exists {
								c.Set("context", dockerfile)
							} else {
								return fmt.Errorf("dockerfile path is required in build mode")
							}
						}

						fmt.Printf("Proceeding with build mode using local Dockerfile: %s\n", c.String("context"))

						dir, err := os.Getwd()
						if err != nil {
							return fmt.Errorf("failed to get current directory: %v", err)
						}

						currentDirectory := filepath.Base(dir)
						fmt.Printf("Current directory: %s\n", currentDirectory)

						sandboxInstance := sandbox.NewSandboxWithLocalDockerfile(sandbox.SandboxOpts{
							DockerContext: c.String("context"),
							Directory:     currentDirectory,
							RemoveAfter:   c.Bool("remove"),
							Ports:         c.IntSlice("ports"),
						})
						sandboxInstance.Start()

					} else {
						fmt.Println("Continuing with available built image")

						dir, err := os.Getwd()
						if err != nil {
							return fmt.Errorf("failed to get current directory: %v", err)
						}

						currentDirectory := filepath.Base(dir)
						fmt.Printf("Current directory: %s\n", currentDirectory)

						if c.String("context") == "" {
							c.Set("context", "ubuntu:latest")
						}
						fmt.Printf("Proceeding using %s image\n", c.String("context"))

						sandboxInstance := sandbox.NewSandboxWithImage(sandbox.SandboxOpts{
							DockerContext: c.String("context"),
							Directory:     currentDirectory,
							RemoveAfter:   c.Bool("remove"),
							Ports:         c.IntSlice("ports"),
						})
						sandboxInstance.Start()
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
