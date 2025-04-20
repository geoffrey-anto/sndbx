package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/geoffrey-anto/sndbx/internal/utils"
)

func GetAvailableEnvironments() []string {
	return []string{
		"ubuntu:latest",
		"debian:latest",
		"alpine:latest",
		"centos:latest",
		"fedora:latest",
	}
}

func CreateContainerWithLocalDockerfile(filename string, directory string) error {
	fmt.Printf("Using %s/%s file to create sndbx environment!\n", directory, filename)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	dir := filepath.Dir(filename)
	dockerfile := filepath.Base(filename)

	buildContext, err := utils.CreateTarContext(dir)

	if err != nil {
		return errors.New("error creating tar build context")
	}

	imageName := fmt.Sprintf("%s-%s", "sndbx", directory)

	sandboxImage, err := cli.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{imageName},
		Remove:     true,
	})

	if err != nil {
		fmt.Printf("%+v\n", err)
		return errors.New("failed to build image")
	}

	io.Copy(os.Stdout, sandboxImage.Body)

	defer sandboxImage.Body.Close()

	fmt.Printf("Created Image %s using local %s\n", sandboxImage.OSType, filename)

	cwd, err := os.Getwd()
	if err != nil {
		return errors.New("failed to get cwd")
	}

	// Create a container from the image
	sandbox_container, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      imageName,
		Tty:        true,
		WorkingDir: "/app",
		Cmd:        []string{"/bin/sh"},
		User:       fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Source: cwd,
				Target: "/app",
			},
		},
	}, nil, nil, imageName)

	if err != nil {
		fmt.Printf("%+v\n", err)
		return errors.New("failed to create container")
	}
	fmt.Printf("Container %s created\n", sandbox_container.ID)

	// Start the container
	if err := cli.ContainerStart(ctx, sandbox_container.ID, container.StartOptions{}); err != nil {
		fmt.Printf("%+v\n", err)
		return errors.New("failed to start container")
	}
	fmt.Printf("Container %s started\n", sandbox_container.ID)

	// Exec interactive command
	execConfig := container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh"},
	}

	execID, err := cli.ContainerExecCreate(ctx, sandbox_container.ID, execConfig)
	if err != nil {
		fmt.Printf("Error creating exec instance: %v\n", err)
		return errors.New("error creating exec instance")
	}

	resp, err := cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{
		Tty: true,
	})
	if err != nil {
		fmt.Printf("Error attaching to exec instance: %v\n", err)
		return errors.New("error attaching to exec instance")
	}
	defer resp.Close()

	// Handle interactive session
	if err := utils.StreamTerminal(resp); err != nil {
		fmt.Printf("Error streaming terminal: %v\n", err)
		return errors.New("error streaming terminal")
	}

	// Remove the container
	if err := cli.ContainerRemove(ctx, sandbox_container.ID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		fmt.Printf("Error removing container: %v\n", err)
		return errors.New("error removing container")
	}

	fmt.Printf("Container %s removed\n", sandbox_container.ID)

	// Remove the image
	if _, err := cli.ImageRemove(ctx, imageName, image.RemoveOptions{}); err != nil {
		fmt.Printf("Error removing image: %v\n", err)
		return errors.New("error removing image")
	}

	fmt.Printf("Image %s removed\n", imageName)

	return nil
}
