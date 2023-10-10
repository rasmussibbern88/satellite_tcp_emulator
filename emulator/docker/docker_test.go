package docker

import (
	"testing"
)

func TestRunContainer(t *testing.T) {
	InstantiateDocker()
	id := RunBackGroundContainer("P7-testing")
	t.Log(id)
	RunCommand(id, "ip route add 172.17.0.3 via 172.17.0.1")
	RunCommand(id, "touch /root/P7.txt")
	Cleanup()
}

func TestAddLink(t *testing.T) {

	InstantiateDocker()
	nodeOneID := RunBackGroundContainer("P7-1-Test")
	nodeTwoID := RunBackGroundContainer("P7-2-Test")
	nodeThreeID := RunBackGroundContainer("P7-3-Test")
	nodeFourID := RunBackGroundContainer("P7-4-Test")

	link1 := linkDetails{
		NodeOneId:   nodeOneID,
		NodeTwoId:   nodeTwoID,
		NetworkName: "P7-link-1",
		Subnet:      "192.168.0.0/24",
		NodeOneIP:   "192.168.0.2",
		NodeTwoIP:   "192.168.0.3",
	}

	link2 := linkDetails{
		NodeOneId:   nodeTwoID,
		NodeTwoId:   nodeThreeID,
		NetworkName: "P7-link-2",
		Subnet:      "192.168.1.0/24",
		NodeOneIP:   "192.168.1.2",
		NodeTwoIP:   "192.168.1.3",
	}

	link3 := linkDetails{
		NodeOneId:   nodeThreeID,
		NodeTwoId:   nodeFourID,
		NetworkName: "P7-link-3",
		Subnet:      "192.168.2.0/24",
		NodeOneIP:   "192.168.2.2",
		NodeTwoIP:   "192.168.2.3",
	}

	CreateLink(link1)
	CreateLink(link2)
	CreateLink(link3)

	Cleanup()
}
