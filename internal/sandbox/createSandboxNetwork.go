package sandbox

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
)

func CreateSandboxNetwork(sandbox *Sandbox, networkName string) (string, error) {
	if networkName == "" {
		return "", fmt.Errorf("network name cannot be empty")
	}

	ctx := context.Background()

	res, err := sandbox.Cli.NetworkCreate(ctx, networkName, network.CreateOptions{
		Labels: map[string]string{
			"created_by": "sndbx",
			"app":        "sndbx",
		},
		Attachable: true,
	})

	if err != nil {
		return "", fmt.Errorf("failed to create network: %w", err)
	}
	if res.ID == "" {
		return "", fmt.Errorf("failed to create network: empty network ID")
	}

	return res.ID, nil
}
