package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/geoffrey-anto/sndbx/internal/utils"
)

type Sandbox struct {
	Cli         *client.Client
	RemoveAfter bool
	Ports       []int
	Plugins     []string
}

type SandboxOpts struct {
	DockerContext string
	Directory     string
	RemoveAfter   bool
	Ports         []int
	Plugins       []string
}

type Plugins struct {
	PluginName  string
	ContainerID string
}

func GetAvailableEnvironments() []string {
	return []string{
		"ubuntu:latest",
		"debian:latest",
		"alpine:latest",
		"centos:latest",
		"fedora:latest",
	}
}

// Function to Check if Image Exists locally or pull it
func CheckIfImageExists(sandbox *Sandbox, ImageName string) (bool, error) {
	ctx := context.Background()

	images, err := sandbox.Cli.ImageList(ctx, image.ListOptions{})

	if err != nil {
		return false, errors.New("failed to list images")
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == ImageName {
				return true, nil
			}
		}
	}

	return false, nil
}

// Function to Pull Image
func PullImage(sandbox *Sandbox, ImageName string) error {
	ctx := context.Background()

	reader, err := sandbox.Cli.ImagePull(ctx, ImageName, image.PullOptions{})
	if err != nil {
		return errors.New("failed to pull image")
	}
	defer reader.Close()

	decoder := json.NewDecoder(reader)

	for decoder.More() {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			continue // skip malformed lines
		}

		progressDetail, ok := msg["progressDetail"].(map[string]interface{})
		if !ok || progressDetail["current"] == nil || progressDetail["total"] == nil {
			continue
		}

		current := int64(progressDetail["current"].(float64))
		total := int64(progressDetail["total"].(float64))
		if total == 0 {
			continue
		}

		percentage := float64(current) / float64(total) * 100
		barWidth := 40
		done := int((percentage / 100) * float64(barWidth))
		bar := "[" + strings.Repeat("=", done) + ">" + strings.Repeat(" ", barWidth-done) + "]"

		fmt.Printf("\r%s %.2f%%", bar, percentage)
	}

	fmt.Println("\n✅ Image pull complete")
	fmt.Printf("Pulled Image %s\n", ImageName)

	return nil
}

// Function to Build Image
func BuildImage(sandbox *Sandbox, DockerContext string, Directory string) (string, error) {
	ctx := context.Background()

	dir := filepath.Dir(DockerContext)
	dockerfile := filepath.Base(DockerContext)

	buildContext, err := utils.CreateTarContext(dir)

	if err != nil {
		return "", errors.New("error creating tar build context")
	}

	imageName := fmt.Sprintf("%s-%s", "sndbx", Directory)

	sandboxImage, err := sandbox.Cli.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{imageName},
		Remove:     true,
	})

	if err != nil {
		fmt.Printf("%+v\n", err)
		return "", errors.New("failed to build image")
	}

	io.Copy(io.Discard, sandboxImage.Body)

	defer sandboxImage.Body.Close()

	fmt.Printf("🏗️  Created Image %s using local %s\n", sandboxImage.OSType, DockerContext)

	return imageName, nil
}

// Function to Start Image
func StartImage(sandbox *Sandbox, DockerContext string, Directory string, ImageName string, NetworkName string) (string, error) {
	ctx := context.Background()

	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.New("failed to get cwd")
	}

	// Define exposed ports
	exposedPorts := make(map[nat.Port]struct{})
	portBindings := make(map[nat.Port][]nat.PortBinding) // Map for port bindings

	for _, port := range sandbox.Ports {
		portStr := fmt.Sprintf("%d/tcp", port)
		exposedPorts[nat.Port(portStr)] = struct{}{}
		portBindings[nat.Port(portStr)] = []nat.PortBinding{
			{HostPort: fmt.Sprintf("%d", port)}, // Mapping container port to host port
		}
	}

	// Create a container from the image
	sandbox_container, err := sandbox.Cli.ContainerCreate(ctx, &container.Config{
		Image:        ImageName,
		Tty:          true,
		WorkingDir:   "/app",
		Cmd:          []string{"/bin/sh"},
		User:         fmt.Sprintf("%d:%d", 0, os.Getgid()),
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Source: cwd,
				Target: "/app",
			},
		},
		NetworkMode:  container.NetworkMode(NetworkName),
		PortBindings: portBindings, // Add the port mappings here
	}, nil, nil, fmt.Sprintf("%s-%s", "sndbx", Directory))

	if err != nil {
		fmt.Printf("%+v\n", err)
		return "", errors.New("failed to create container")
	}

	fmt.Printf("📦 Container %s created\n", sandbox_container.ID)

	return sandbox_container.ID, nil
}

func StartPluginImage(sandbox *Sandbox, PluginImageName string, NetworkName string) (string, error) {
	ctx := context.Background()

	// Create a container for the plugin image; Eg: postgres as "sndbx-postgres"
	sandbox_container, err := sandbox.Cli.ContainerCreate(ctx, &container.Config{
		Image: PluginImageName,
	}, &container.HostConfig{
		NetworkMode: container.NetworkMode(NetworkName),
	}, nil, nil, fmt.Sprintf("%s-%s", "sndbx", PluginImageName))

	if err != nil {
		err = PullImage(sandbox, PluginImageName)
		if err != nil {
			fmt.Printf("%+v\n", err)
			return "", errors.New("failed to pull plugin image")
		}

		// Retry creating the container after pulling the image
		sandbox_container, err = sandbox.Cli.ContainerCreate(ctx, &container.Config{
			Image: PluginImageName,
		}, &container.HostConfig{
			NetworkMode: container.NetworkMode(NetworkName),
		}, nil, nil, fmt.Sprintf("%s-%s", "sndbx", PluginImageName))

		// Check for errors again
		// If it fails again, return the error
		if err != nil {
			fmt.Printf("%+v\n", err)
			return "", errors.New("failed to create plugin container")
		}
	}

	fmt.Printf("📦 Plugin Sidecar Container %s created\n", sandbox_container.ID)

	return sandbox_container.ID, nil
}

// Function to Create and Attach Exec
func CreateAndAttachExec(sandbox *Sandbox, DockerContext string, Directory string, ContainerID string) error {
	ctx := context.Background()

	// Start the container
	if err := sandbox.Cli.ContainerStart(ctx, ContainerID, container.StartOptions{}); err != nil {
		fmt.Printf("%+v\n", err)
		return errors.New("failed to start container")
	}
	fmt.Printf("▶️  Container %s started\n", ContainerID)

	// Exec interactive command
	execConfig := container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh"},
	}

	execID, err := sandbox.Cli.ContainerExecCreate(ctx, ContainerID, execConfig)
	if err != nil {
		fmt.Printf("Error creating exec instance: %v\n", err)
		return errors.New("error creating exec instance")
	}

	resp, err := sandbox.Cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{
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

	return nil
}

// Function to Cleanup Container
func CleanupContainer(sandbox *Sandbox, DockerContext string, Directory string, ImageName string, ContainerID string, plugins []Plugins, NetworkName string) error {
	ctx := context.Background()

	if err := sandbox.Cli.ContainerRemove(ctx, ContainerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		fmt.Printf("Error removing container: %v\n", err)
		return errors.New("error removing container")
	}

	fmt.Printf("🗑️  Container %s removed\n", ContainerID)

	// Remove the image
	if _, err := sandbox.Cli.ImageRemove(ctx, ImageName, image.RemoveOptions{}); err != nil {
		fmt.Printf("Error removing image: %v\n", err)
		return errors.New("error removing image")
	}

	// Remove the plugins
	for _, plugin := range plugins {
		if err := sandbox.Cli.ContainerRemove(ctx, plugin.ContainerID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			fmt.Printf("Error removing plugin container: %v\n", err)
			return errors.New("error removing plugin container")
		}
		fmt.Printf("🗑️  Plugin Sidecar Container %s removed\n", plugin.ContainerID)
	}
	// Remove the network
	if err := sandbox.Cli.NetworkRemove(ctx, NetworkName); err != nil {
		fmt.Printf("Error removing network: %v\n", err)
		return errors.New("error removing network")
	}

	fmt.Printf("🗑️  Image %s removed\n", ImageName)

	return nil
}
