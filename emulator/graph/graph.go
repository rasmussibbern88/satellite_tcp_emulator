package graph

import (
	"errors"
	"project/space"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yourbasic/graph"
)

//var (
// g *graph.Mutable
// v int
//)
// const hopPenalty float32 =

func InstantiateGraph(vertices int) (g *graph.Mutable) {
	g = graph.New(vertices)
	return g
}

// AddBothCost inserts edges with cost between node1 and node2. It overwrites the previous costs if these edges already exist.
func AddBothCost(g *graph.Mutable, v int, node1 int, node2 int, linkCost int64) error {
	if g == nil {
		return errors.New("there is no graph instantiated")
	}
	if node1 > v-1 || node2 > v-1 {
		return errors.New("out of range")
	}
	g.AddBothCost(node1, node2, linkCost)

	return nil
}

func GetShortestPath(g *graph.Mutable, v int, node1 int, node2 int) (path []int, dist int64, e error) {
	if g == nil {
		return nil, 0, errors.New("there is no graph instantiated")
	}
	if node1 > v-1 || node2 > v-1 {
		return nil, 0, errors.New("out of range")
	}
	path, dist = graph.ShortestPath(g, node1, node2)
	return path, dist, nil
}

func SetupGraphSatelliteEdges(g *graph.Mutable, index int, satdata []space.OrbitalData, maxFSODistance float64) {
	for node1, satFrom := range satdata {
		// fmt.Println("%i", satFrom.Time_steps[index].Day())
		for node2, satTo := range satdata {
			if node1 == node2 {
				continue
			}
			var err error
			// relative_velocity := satFrom.Position[index].Sub(satTo.Position[index])
			// speed := relative_velocity.Speed() > 10e3

			if space.Reachable(satFrom.Position[index], satTo.Position[index], maxFSODistance) {

				distance := satFrom.Position[index].Distance(satTo.Position[index]) // Refactoring space would allow on less distance computation per link
				cost := space.Latency(distance) * 1000000
				err = AddBothCost(g, len(satdata), node1, node2, int64(cost))

				//TODO: add network to docker to simulate link
			} else {
				err = AddBothCost(g, len(satdata), node1, node2, -1)
			}
			if err != nil {
				log.Error().Int("satFrom", satFrom.SatelliteId).Int("satTo", satTo.SatelliteId).Err(err).Msg("Error in adding edge")
			}
		}
	}
}

func SetupGraphGroundStationEdges(g *graph.Mutable, index int, simulationTime time.Time, satdata []space.OrbitalData, gsdata []space.GroundStation, maxFSODistance float64) {
	for gsid, gs := range gsdata {
		if !gs.IsAP {
			continue
		}
		for node1, sat := range satdata {
			visible, distance := space.SatelliteVisible(&gs, sat.LatLong[index])
			if visible {
				log.Debug().Bool("visible", visible).Float64("distance", distance).Msg("satellite visibility")
			}
			if !visible || distance > 1500 {
				err := AddBothCost(g, len(gsdata)+len(satdata), len(satdata)+gsid, node1, -1)
				if err != nil {
					log.Error().Err(err).Str("gsname", gs.Title).Msg("failed to add -1 path to graph")
				}
				continue
			}
			cost := space.Latency(float64(distance)) * 1000000
			err := AddBothCost(g, len(gsdata)+len(satdata), len(satdata)+gsid, node1, int64(cost))
			if err != nil {
				log.Error().Err(err).Str("gsname", gs.Title).Msg("failed to add cost path to graph")
			}
			log.Info().Bool("visible", visible).Float64("distance", distance).Str("gsname", gs.Title).Msg("adding GS link")
		}
	}
}

func SetupGraphAccessPointEdges(g *graph.Mutable, graphSize int, gsdata []space.GroundStation, maxAPDistance float64) {
	graphOffset := graphSize - len(gsdata)
	for gs1id, gs1 := range gsdata {
		if !gs1.IsAP { // Compare all Access Points
			continue
		}
		for gs2id, gs2 := range gsdata {
			if gs1id == gs2id {
				continue // Skip connections to itself
			}
			if gs2.IsAP { // To non Access Points
				continue
			}

			visible, distance := space.AccessPointVisible(&gs1, &gs2, maxAPDistance)
			if visible {
				log.Debug().Bool("visible", visible).Float64("distance", distance).Msg("Access Point In Range")
			}

			if !visible {
				err := AddBothCost(g, graphSize, graphOffset+gs1id, graphOffset+gs2id, -1)
				if err != nil {
					log.Error().Err(err).Str("gs1name", gs1.Title).Str("gs2name", gs2.Title).Msg("failed to add -1 path to graph")
				}
				continue
			}
			cost := space.Latency(float64(distance)) * 1000000
			err := AddBothCost(g, graphSize, graphOffset+gs1id, graphOffset+gs2id, int64(cost))
			if err != nil {
				log.Error().Err(err).Str("gs1name", gs1.Title).Str("gs2name", gs2.Title).Msg("failed to add cost path to graph")
			}
			log.Info().Bool("visible", visible).Float64("distance", distance).Str("gs1name", gs1.Title).Str("gs2name", gs2.Title).Msg("adding AP link")
		}
	}
}

func IsPathInGraph(g *graph.Mutable, path []int) (found bool, e error) {
	if g == nil {
		return false, errors.New("there is no graph instantiated")
	}
	for i := 0; i < len(path)-1; i++ {
		val := g.Cost(path[i], path[i+1])
		if val == -1 {
			return false, nil
		}
	}
	return true, nil
}
