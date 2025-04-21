package sandbox

import (
	"os"

	"github.com/docker/docker/client"
)

type LocalImageSandbox struct {
	LocalDockerfile string
	Directory       string
	*Sandbox
}

func NewSandboxWithLocalDockerfile(sandbxOpts SandboxOpts) *LocalImageSandbox {
	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		panic(err)
	}

	filesInCurrentDirectory, err := os.ReadDir(".")

	if err != nil {
		panic("Cannot read current file path! Please check permissions\n")
	}

	for _, fileInCurrentDirectory := range filesInCurrentDirectory {
		if fileInCurrentDirectory.Name() == sandbxOpts.DockerContext {
			return &LocalImageSandbox{
				LocalDockerfile: sandbxOpts.DockerContext,
				Directory:       sandbxOpts.Directory,
				Sandbox: &Sandbox{
					Cli: cli,
				},
			}
		}
	}

	panic("file not found")
}

func (sandbox *LocalImageSandbox) Start() {
	imageName, err := BuildImage(sandbox.Sandbox, sandbox.LocalDockerfile, sandbox.Directory)

	if err != nil {
		panic("failed to build image")
	}

	containerID, err := StartImage(sandbox.Sandbox, sandbox.LocalDockerfile, sandbox.Directory, imageName)

	if err != nil {
		panic("failed to start container")
	}

	err = CreateAndAttachExec(sandbox.Sandbox, sandbox.LocalDockerfile, sandbox.Directory, containerID)

	if err != nil {
		panic("failed to attach exec")
	}

	err = CleanupContainer(sandbox.Sandbox, sandbox.LocalDockerfile, sandbox.Directory, imageName, containerID)

	if err != nil {
		panic("failed to cleanup container")
	}
}
