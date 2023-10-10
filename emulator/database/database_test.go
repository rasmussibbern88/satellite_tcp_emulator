package database

import (
	"project/space"
	"testing"
)

// Verify parquet data
func TestParseSatelliteData(t *testing.T) {
	satData := LoadSatellitePositions("../constellation.parquet")
	earthPosition := space.Vector3{X: 0, Y: 0, Z: 0}

	for _, satellite := range satData {
		for i := 1; i < len(satellite.Position); i++ {
			distance := satellite.Position[i].Distance(satellite.Position[i-1])

			if distance > 12 /* km/s */ *15 /* s */ || distance < 6*15 {
				t.Log(distance, i, satellite.SatelliteId)
				t.FailNow()
			}
			earth_offset := earthPosition.Distance(satellite.Position[i])
			if earth_offset > 6371+1200+100 || earth_offset < 6371+1200-100 {
				t.Fail()
				t.Log("satellite not in correct orbit", satellite.SatelliteId)
			}
		}
	}

}
