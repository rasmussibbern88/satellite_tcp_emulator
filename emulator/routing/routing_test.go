package routing

import (
	"testing"
)

func TestRouting(t *testing.T) {
	var nodes []int = []int{
		0, 32, 75, 79, 94,
	}
	forward, reverse := RouteTables(nodes)
	t.Log(forward, reverse)
	// InstantiateDocker()
	// id := RunBackGroundContainer("P7-testing")
	// t.Log(id)
	// RunCommand(id, "ip route add 172.17.0.3 via 172.17.0.1")
	// RunCommand(id, "touch /root/P7.txt")
	// Cleanup()
}
