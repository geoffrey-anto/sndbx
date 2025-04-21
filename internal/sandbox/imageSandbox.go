package sandbox

import "github.com/docker/docker/client"

type RemoteImageSandbox struct {
	ImageName string
	Directory string
	*Sandbox
}

func NewSandboxWithImage(sandbxOpts SandboxOpts) *RemoteImageSandbox {
	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		panic(err)
	}

	if sandbxOpts.DockerContext == "" {
		panic("provide image to be used")
	}

	return &RemoteImageSandbox{
		ImageName: sandbxOpts.DockerContext,
		Directory: sandbxOpts.Directory,
		Sandbox: &Sandbox{
			Cli: cli,
		},
	}
}

func (sandbox *RemoteImageSandbox) Start() {
	exists, err := CheckIfImageExists(sandbox.Sandbox, sandbox.ImageName)
	if err != nil {
		panic("failed to check if image exists")
	}

	if !exists {
		err = PullImage(sandbox.Sandbox, sandbox.ImageName)
		if err != nil {
			panic("failed to pull image")
		}
	}

	containerID, err := StartImage(sandbox.Sandbox, sandbox.ImageName, sandbox.Directory, sandbox.ImageName)

	if err != nil {
		panic("failed to start container")
	}

	err = CreateAndAttachExec(sandbox.Sandbox, sandbox.ImageName, sandbox.Directory, containerID)

	if err != nil {
		panic("failed to attach exec")
	}

	err = CleanupContainer(sandbox.Sandbox, sandbox.ImageName, sandbox.Directory, sandbox.ImageName, containerID)

	if err != nil {
		panic("failed to cleanup container")
	}
}
