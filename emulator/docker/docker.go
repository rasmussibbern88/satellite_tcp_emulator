package docker

import (
	"context"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

var (
	c         *client.Client
	imageName = "satellite:latest"
)

func InstantiateDocker() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	c = cli
	if err != nil {
		panic(err)
	}
	_, err = c.Info(ctx)
	if err != nil {
		panic(err)
	}
	// out, err := c.ImagePull(ctx, imageName, types.ImagePullOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// defer out.Close()
	// io.Copy(os.Stdout, out)
}

func RunBackGroundContainer(containerName string) string {
	ctx := context.Background()
	// c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// if err != nil {
	// 	panic(err)
	// }
	log.Info().Str("name", containerName).Msg("Creating container")
	resp, err := c.ContainerCreate(ctx, &container.Config{
		Image:    imageName,
		Hostname: containerName,
	},
		&container.HostConfig{
			Privileged: true,
			CapAdd:     strslice.StrSlice([]string{"NET_ADMIN"})},
		nil, nil, containerName)
	if err != nil {
		log.Error().Err(err)
		return ""
	}
	if err := c.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Error().Err(err)
		return ""
	}

	log.Info().Str("ID", resp.ID).Msg("Started container with ID")
	return resp.ID
}

type linkDetails struct {
	NetworkName string `parquet:"network_name"`
	Subnet      string `parquet:"subnet"`
	NodeOneId   string `parquet:"node_one_id"`
	NodeOneIP   string `parquet:"node_one_ip"`
	NodeTwoId   string `parquet:"node_two_id"`
	NodeTwoIP   string `parquet:"node_two_ip"`
}

func createNetwork(name string, subnet string) string {
	ctx := context.Background()

	newnetwork := types.NetworkCreate{IPAM: &network.IPAM{
		Driver: "default",
		Config: []network.IPAMConfig{{
			Subnet: subnet,
		}},
	}}

	res, err := c.NetworkCreate(ctx, name, newnetwork)
	if err != nil {
		log.Error().Err(err)
		return ""
	}
	return res.ID
}

func addContainerToNetwork(containerId string, networkId string, IPAddress string) {
	ctx := context.Background()
	networkConf := network.EndpointSettings{
		IPAddress: IPAddress,
	}
	c.NetworkConnect(ctx, networkId, containerId, &networkConf)
}

func CreateLink(details linkDetails) {
	networkId := createNetwork(details.NetworkName, details.Subnet)

	addContainerToNetwork(details.NodeOneId, networkId, details.NodeOneIP)
	addContainerToNetwork(details.NodeTwoId, networkId, details.NodeTwoIP)
}

func RunCommand(nodeID string, command string) error {
	ctx := context.Background()
	cmd := strings.Split(command, " ")
	log.Info().Str("node", nodeID).Strs("command", cmd).Msg("Running command")
	execId, err := c.ContainerExecCreate(ctx, nodeID, types.ExecConfig{
		Privileged: true,
		User:       "root",
		Cmd:        cmd,
	})
	if err != nil {
		log.Error().Err(err)
		return err
	}
	err = c.ContainerExecStart(ctx, execId.ID, types.ExecStartCheck{})
	if err != nil {
		log.Error().Err(err)
		return err
	}
	return nil
}

func Cleanup() {
	ctx := context.Background()
	log.Info().Msg("Cleaning up")
	containers, err := c.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if container.Image == imageName {
			log.Info().Str("ID", container.ID[:10]).Msg("Stopping container")

			timeoutTime := time.Millisecond * 1
			if err := c.ContainerStop(ctx, container.ID, &timeoutTime); err != nil {
				panic(err)
			}
			log.Info().Str("ID", container.ID[:10]).Msg("Removing container")

			if err := c.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
				panic(err)
			}
		}
	}

	networks, err := c.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for _, network := range networks {
		if strings.Contains(network.Name, "P7") {
			log.Info().Str("ID", network.ID).Msg("Removing network")
			if err := c.NetworkRemove(ctx, network.ID); err != nil {
				panic(err)
			}
		}
	}

}
