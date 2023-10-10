package podman

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/containers/common/libnetwork/types"
	dockertypes "github.com/docker/docker/api/types"

	// "github.com/containers/libpod/libpod/define"
	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/api/handlers"
	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/bindings/network"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/rs/zerolog/log"
)

var (
	ctx                   context.Context
	SatelliteRawImage     = "registry.hub.docker.com/mulvadt/satellite:latest"
	GroundStationRawImage = "registry.hub.docker.com/mulvadt/groundstation:latest"
	// rawImage = "localhost/satellite"
)

func InitPodman() {
	log.Info().Msg("Podman is now initializing")

	// Get Podman socket location
	sock_dir := os.Getenv("XDG_RUNTIME_DIR")
	if sock_dir == "" {
		sock_dir = "/var/run"
	}
	socket := "unix:" + sock_dir + "/podman/podman.sock"

	// Connect to Podman socket
	contex, err := bindings.NewConnection(context.Background(), socket)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ctx = contex

	if os.Getenv("FORCE_PULL") == "TRUE" {
		_, err = images.Pull(ctx, SatelliteRawImage, &images.PullOptions{})
		if err != nil {
			log.Err(err).Msg(err.Error())
		}

		_, err = images.Pull(ctx, GroundStationRawImage, &images.PullOptions{})
		if err != nil {
			log.Err(err).Msg(err.Error())
		}
	}

}

func ListContainers() {
	// List images
	imageSummary, err := images.List(ctx, &images.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list images")
		os.Exit(1)
	}
	var names []string
	for _, i := range imageSummary {
		names = append(names, i.RepoTags...)
	}
	log.Info().Strs("Images", names).Msg("List of images")

	// Container list
	var latestContainers = 1
	containerLatestList, err := containers.List(ctx, &containers.ListOptions{
		Last: &latestContainers,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list CONTAINERS")
		os.Exit(1)
	}
	log.Info().Int("Containers", len(containerLatestList)).Strs("names", containerLatestList[0].Names).Msg("List of containers")
}

func CreateRunContainer(containerName string, usetty bool, rawImage string) string {
	// Container create
	var capadd []string
	capadd = append(capadd, "NET_ADMIN")
	capadd = append(capadd, "NET_RAW")
	s := specgen.NewSpecGenerator(rawImage, false)
	s.Terminal = usetty
	s.CapAdd = capadd
	// s.NetNS = specgen.Namespace{ //Added 16/11
	// 	// NSMode: specgen.NoNetwork,
	// 	NSMode: specgen.Private,
	// }
	r, err := containers.CreateWithSpec(ctx, s, &containers.CreateOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create container")
		// os.Exit(1)
	}

	err = containers.Rename(ctx, r.ID, &containers.RenameOptions{Name: &containerName})
	if err != nil {
		log.Error().Err(err).Msg("Failed to name container")
	}

	// Container start
	log.Info().Str("Name", containerName).Msg("Starting sattelite container...")
	err = containers.Start(ctx, r.ID, &containers.StartOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to start container")

	}

	_, err = containers.Wait(ctx, r.ID, &containers.WaitOptions{
		Condition: []define.ContainerStatus{define.ContainerStateRunning},
	})
	if err != nil {
		fmt.Println(err)
	}
	network.Disconnect(ctx, "podman", containerName, &network.DisconnectOptions{})
	return containerName
}

type LinkDetails struct {
	NetworkName string `parquet:"network_name"`
	Subnet      string `parquet:"subnet"`
	NodeOneId   string `parquet:"node_one_id"`
	NodeOneIP   string `parquet:"node_one_ip"`
	NodeTwoId   string `parquet:"node_two_id"`
	NodeTwoIP   string `parquet:"node_two_ip"`
}

func SetupLink(linkDetails LinkDetails) {
	var subnets []types.Subnet
	// log.Debug().Str("subnet", linkDetails.Subnet).Msg("Subnet string")
	ip, ipnet, err := net.ParseCIDR(linkDetails.Subnet)
	if err != nil {
		log.Error().Err(err).Interface("linkDetails", linkDetails).Msg("Failed to parse subnet")
		return
	}

	subnets = append(subnets, types.Subnet{
		Subnet:     types.IPNet{IPNet: *ipnet},
		Gateway:    ip,
		LeaseRange: &types.LeaseRange{StartIP: net.ParseIP(linkDetails.NodeOneIP), EndIP: net.ParseIP(linkDetails.NodeTwoIP)},
	})
	// subnets = append(subnets, types.Subnet{
	// 	Subnet:     types.IPNet{IPNet: net.IPNet{IP: net.ParseIP(linkDetails.Subnet), Mask: net.IPv4Mask(255, 255, 255, 248)}},
	// 	Gateway:    net.ParseIP(linkDetails.Subnet),
	// 	LeaseRange: &types.LeaseRange{},
	// })

	networkSettings := types.Network{Name: linkDetails.NetworkName, Subnets: subnets, Internal: true}

	cn, err := network.Create(ctx, &networkSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create network")
		return
	}

	network.Disconnect(ctx, "podman", linkDetails.NodeOneId, &network.DisconnectOptions{})
	network.Disconnect(ctx, "podman", linkDetails.NodeTwoId, &network.DisconnectOptions{})

	connectContainerToNetwork(cn.Name, linkDetails.NodeOneIP, linkDetails.NodeOneId, linkDetails.NodeTwoId)
	connectContainerToNetwork(cn.Name, linkDetails.NodeTwoIP, linkDetails.NodeTwoId, linkDetails.NodeOneId)

}

func TearDownLink(linkDetails LinkDetails) {
	time.Sleep(5 * time.Second) // if aggregated buffer size is 50MB it takes 4 seconds to empty
	// var forceSetting bool = false
	// var timeoutSetting uint = 0
	err := network.Disconnect(ctx, linkDetails.NetworkName, linkDetails.NodeOneId, &network.DisconnectOptions{})
	if err != nil {
		log.Fatal().Str("containerName", linkDetails.NodeOneId).Str("networkName", linkDetails.NetworkName).Err(err).Msg("failed to disconnect container from network")
	}
	err = network.Disconnect(ctx, linkDetails.NetworkName, linkDetails.NodeTwoId, &network.DisconnectOptions{})
	if err != nil {
		log.Fatal().Str("containerName", linkDetails.NodeTwoId).Str("networkName", linkDetails.NetworkName).Err(err).Msg("failed to disconnect container from network")
	}
	report, err := network.Remove(ctx, linkDetails.NetworkName, &network.RemoveOptions{
		// Force:   &forceSetting,
		// Timeout: &timeoutSetting,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove network")
		return
	}
	log.Debug().Interface("report", report).Msg("Removed Network")
}

func connectContainerToNetwork(cnname string, cip string, cid string, ifname string) {
	var ip []net.IP
	ip = append(ip, net.ParseIP(cip))
	// log.Debug().Str("cip", cip).Str("cid", cid).Msg("Connecting to network")
	err := network.Connect(ctx, cnname, cid, &types.PerNetworkOptions{StaticIPs: ip, InterfaceName: ifname})
	if err != nil {
		log.Error().Str("ifname", ifname).Err(err).Msg("Failed to connect network")
		return
	}
}

func RunCommand(nodeID string, command string) error {
	cmd := strings.Split(command, " ")
	execId, err := containers.ExecCreate(ctx, nodeID, &handlers.ExecCreateConfig{
		ExecConfig: dockertypes.ExecConfig{
			Privileged: true,
			User:       "root",
			Cmd:        cmd,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create exec command")
		return err
	}
	err = containers.ExecStart(ctx, execId, &containers.ExecStartOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to start exec command")
		return err
	}
	return nil
}

func Cleanup() {
	containerLatestList, err := containers.List(ctx, &containers.ListOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// log.Debug().Str("name", containerLatestList[0].Names[0]).Msg("Latest container is")

	duration := uint(1)
	forceBool := true
	wg := new(sync.WaitGroup)
	wg.Add(len(containerLatestList))
	for _, c := range containerLatestList {
		log.Info().Str("Name", c.Names[0]).Str("ImageId ", c.ImageID).Msg("Cleaning up container")
		// log.Warn().Str("image name", c.Image).Msg("satellite image name")
		if c.Image == SatelliteRawImage || c.Image == GroundStationRawImage {
			go func(ctx context.Context, id string) {
				defer wg.Done()
				duration := uint(1)
				forceBool := true
				log.Info().Msg("Stopping the container...")
				//Container stop
				err := containers.Stop(ctx, id, &containers.StopOptions{Ignore: &forceBool, Timeout: &duration})
				if err != nil {
					log.Error().Err(err).Msg("Error stopping container")
				}
				_, err = containers.Remove(ctx, id, &containers.RemoveOptions{})
				if err != nil {
					log.Error().Err(err).Msg("Error removing container")
				}
			}(ctx, c.ID)
		}
	}
	wg.Wait()

	networks, err := network.List(ctx, &network.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Could not get list of networks")
	}
	for _, nw := range networks {
		log.Info().Str("name", nw.Name).Msg("removing network...")
		if strings.Contains(nw.Name, "P7") {
			_, err := network.Remove(ctx, nw.ID, &network.RemoveOptions{Force: &forceBool, Timeout: &duration})
			if err != nil {
				log.Error().Err(err).Msg("Error removing network")

			}

		}
	}
}
