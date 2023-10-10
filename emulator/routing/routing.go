package routing

import (
	"fmt"
	"project/podman"
	"strconv"
)

var (
	LINKS map[string]podman.LinkDetails
)

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func linkNameFromNodeId(node1, node2 int) (string, bool) {
	if node1 == node2 {
		panic("aaaaaaaaaaaaa!")
	}
	firstNode := min(node1, node2)
	secondNode := max(node1, node2)
	return "S" + strconv.Itoa(firstNode) + "-S" + strconv.Itoa(secondNode), firstNode != node1
}

// nodes = path.
func RouteTables(nodes []int) (map[int]string, map[int]string) { //nodes is path
	commands := make(map[int]string)
	reversecommands := make(map[int]string)
	linkid, ss := linkNameFromNodeId(nodes[len(nodes)-2], nodes[len(nodes)-1])
	linkDetails := LINKS[linkid]
	destinationIP := linkDetails.NodeOneIP
	if ss {
		destinationIP = linkDetails.NodeTwoIP
	}
	for i, node := range nodes { // Forward Routing
		if i == len(nodes)-2 {
			break
		}
		linkid2, swapped := linkNameFromNodeId(nodes[i], nodes[i+1])
		link := LINKS[linkid2]
		// fmt.Println(link)
		nexthopIP := link.NodeOneIP
		if swapped {
			nexthopIP = link.NodeTwoIP
		}
		commands[node] = ipRouteVia(destinationIP, nexthopIP)
	}
	// 0  1   2   3]
	//[0, 63, 67, 94]
	linkid3, s := linkNameFromNodeId(nodes[0], nodes[1])
	link := LINKS[linkid3]
	destinationIP = link.NodeTwoIP
	if s {
		destinationIP = link.NodeOneIP
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		if i <= 1 {
			break
		}
		linkid4, swp := linkNameFromNodeId(nodes[i-1], nodes[i])
		link := LINKS[linkid4]
		// fmt.Println(link)

		nexthopIP := link.NodeTwoIP
		if swp {
			nexthopIP = link.NodeOneIP
		}
		reversecommands[nodes[i]] = ipRouteVia(destinationIP, nexthopIP)
	}
	return commands, reversecommands
}

func ipRouteVia(destinationIP, nexthopIP string) string {
	return fmt.Sprintf("ip route replace %s via %s", destinationIP, nexthopIP)
}
