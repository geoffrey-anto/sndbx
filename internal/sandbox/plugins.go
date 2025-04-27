package sandbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

func CreateAndAttachPlugins(sandbox *Sandbox, plugins []string, networkName string) ([]Plugins, error) {
	attachedPlugins := make([]Plugins, 0)
	ctx := context.Background()

	for _, plugin := range plugins {
		fmt.Printf("🔌 Creating plugin container for %s\n", plugin)

		pluginContainerId, err := StartPluginImage(sandbox, plugin, networkName)
		if err != nil {
			fmt.Printf("Error creating plugin container: %s\n", err)
			return nil, err
		}

		if err := sandbox.Cli.ContainerStart(ctx, pluginContainerId, container.StartOptions{}); err != nil {
			fmt.Printf("%+v\n", err)
			return nil, errors.New("failed to start container")
		}
		fmt.Printf("▶️  Container %s started\n", pluginContainerId)

		attachedPlugins = append(attachedPlugins, Plugins{
			PluginName:  plugin,
			ContainerID: pluginContainerId,
		})
	}

	return attachedPlugins, nil
}
