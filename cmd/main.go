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
					&cli.StringSliceFlag{
						Name:  "plugins",
						Usage: "List of Containers to be attached to the sandbox via network bridge",
						Value: cli.NewStringSlice(""),
					},
					&cli.StringFlag{
						Name:    "env",
						Aliases: []string{"e"},
						Usage:   "Environment variables to set in the sandbox",
						Value:   "",
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
							Plugins:       c.StringSlice("plugins"),
							EnvFile:       c.String("env"),
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
							return fmt.Errorf("docker image name is required, Eg: %+v", sandbox.GetAvailableEnvironments())
						}
						fmt.Printf("Proceeding using %s image\n", c.String("context"))

						sandboxInstance := sandbox.NewSandboxWithImage(sandbox.SandboxOpts{
							DockerContext: c.String("context"),
							Directory:     currentDirectory,
							RemoveAfter:   c.Bool("remove"),
							Ports:         c.IntSlice("ports"),
							Plugins:       c.StringSlice("plugins"),
							EnvFile:       c.String("env"),
						})
						sandboxInstance.Start()
					}
					return nil
				},
			},
			{
				Name:  "clear",
				Usage: "Clear all sandboxes",
				Action: func(c *cli.Context) error {
					sandboxClient, err := sandbox.NewSandboxClient()

					if err != nil {
						return fmt.Errorf("failed to create sandbox client: %v", err)
					}

					sandboxClient.Clear()

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
