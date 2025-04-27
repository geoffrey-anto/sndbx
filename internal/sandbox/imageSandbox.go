package sandbox

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/geoffrey-anto/sndbx/internal/utils"
)

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

	env := utils.ParseEnv(sandbxOpts.EnvFile)

	return &RemoteImageSandbox{
		ImageName: sandbxOpts.DockerContext,
		Directory: sandbxOpts.Directory,
		Sandbox: &Sandbox{
			Cli:         cli,
			RemoveAfter: sandbxOpts.RemoveAfter,
			Ports:       sandbxOpts.Ports,
			Plugins:     sandbxOpts.Plugins,
			Envs:        env,
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

	networkId, err := CreateSandboxNetwork(sandbox.Sandbox, "sandbox_network")
	if err != nil {
		panic("failed to create network")
	}

	plugins, err := CreateAndAttachPlugins(sandbox.Sandbox, sandbox.Plugins, networkId)
	if err != nil {
		panic("failed to create and attach plugins")
	}

	containerID, err := StartImage(sandbox.Sandbox, sandbox.ImageName, sandbox.Directory, sandbox.ImageName, networkId)

	if err != nil {
		panic("failed to start container")
	}

	err = CreateAndAttachExec(sandbox.Sandbox, sandbox.ImageName, sandbox.Directory, containerID)

	if err != nil {
		panic("failed to attach exec")
	}

	if !sandbox.RemoveAfter {
		fmt.Printf("Container running with ID: %s\n", containerID)
		fmt.Printf("To stop/remove the container, run:\n")
		fmt.Printf("docker stop/rm %s\n", containerID)

		fmt.Printf("To stop/remove the plugin containers, run:\n")
		for _, plugin := range plugins {
			fmt.Printf("docker stop/rm %s\n", plugin.ContainerID)
		}
		fmt.Printf("To stop/remove the network, run:\n")
		fmt.Printf("docker network rm %s\n", networkId)
		return
	}

	err = CleanupContainer(sandbox.Sandbox, sandbox.ImageName, sandbox.Directory, sandbox.ImageName, containerID, plugins, networkId)

	if err != nil {
		panic("failed to cleanup container")
	}
}
