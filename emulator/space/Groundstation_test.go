package space

import (
	"fmt"
	"log"
	"os"
	"project/tle"
	"testing"
	"time"

	gosat "github.com/joshuaferrara/go-satellite"
)

func TestGroundStation(t *testing.T) {
	Aalborg := GroundStationGen("Aalborg", 0, 57.014077, 9.986923, true)
	Tokyo := GroundStationGen("Tokyo", 1, 35.678892, 139.768596, true)

	d, isin := CalculateDistance(Aalborg.Latlong, Tokyo.Latlong, 300)
	altitude := 0.0
	jday_0 := gosat.GSTimeFromDate(2022, 12, 12, 12, 12, 12)
	for i := 0; i < 22; i++ {
		jday_timestamp := jday_0 + float64(10*i)
		aalborgPos := gosat.LLAToECI(Aalborg.Latlong.asGosatLatLone(), altitude, jday_timestamp)
		tokyoPos := gosat.LLAToECI(Tokyo.Latlong.asGosatLatLone(), altitude, jday_timestamp)
		// T, A := gstation(Aalborg, Tokyo, tspamp, 10)
		fmt.Printf("%v: %v\n", aalborgPos, tokyoPos)
		fmt.Printf("Distance in km: %f\nIs point in circle : %t\n", d, isin)
	}
}

func TestGroundStationLoad(t *testing.T) {
	groundstations, err := LoadGroundStations("../groundstations.txt")
	if err != nil {
		t.Error(err)
	}
	t.Log(groundstations)
	if len(groundstations) == 0 {
		t.Fail()
	}
}

func TestCalculateDistance(t *testing.T) {
	startingPoint := LatLong{
		Latitude:  57.02025,
		Longitude: 10.00492,
	}
	TenAway := LatLong{
		Latitude:  57.11020,
		Longitude: 10.00562,
	}
	testDistance := 15
	d, incircle := CalculateDistance(startingPoint, TenAway, 9)
	t.Log(d, incircle)
	if incircle {
		t.Errorf("point is not in circle r=%d but was calculated as distance %f ", testDistance, d)
	}
}

func TestSatelliteProjection(t *testing.T) {
	st := gosat.NewSpacetrack("rfrede19@student.aau.dk", os.Getenv("SPACETRACK_PASS"))

	startTime := time.Now().UTC()
	log.Println(startTime)
	aausat4_id := []int{41460}
	satellites := tle.GetTLEfromNoradIDs(st, aausat4_id)

	sat_data := GetSatData(satellites, []int{1}, startTime, 1*time.Second, 10*time.Second)
	satellite_data := sat_data[0]

	ll := satellite_data.LatLong[0]
	// 25.416750724987093, -54.989271422423016
	startingPoint := LatLong{
		Latitude:  -65.416750724987093,
		Longitude: -82.989271422423016,
	}

	log.Print(ll.Latitude, ll.Longitude)
	d, incircle := CalculateDistance(startingPoint, ll, 1000)
	if !incircle {
		t.Error("not in circle, distance", d)
	}
	// Test passed
}

func TestSatelliteLookAngles(t *testing.T) {
	st := gosat.NewSpacetrack("rfrede19@student.aau.dk", os.Getenv("SPACETRACK_PASS"))

	startTime := time.Now().UTC()
	log.Println(startTime)
	aausat4_id := []int{41460}
	satellites := tle.GetTLEfromNoradIDs(st, aausat4_id)

	sat_data := GetSatData(satellites, []int{1}, startTime, 1*time.Second, 10*time.Second)
	satellite_data := sat_data[0]

	ll := satellite_data.LatLong[0]
	// 25.416750724987093, -54.989271422423016
	// startingPoint := LatLong{
	// 	Latitude:  55.6167,
	// 	Longitude: 12.650,
	// }
	startingPoint := LatLong{
		Latitude:  30,
		Longitude: -151,
	}

	distance, incircle := CalculateDistance(startingPoint, ll, 1500)

	t.Log(distance, incircle)
	if !incircle {
		t.Fail()
	}

	// gs := GroundStationGen("punk", 0, startingPoint.Latitude, startingPoint.Longitude, false)
	// log.Print(ll.Latitude, ll.Longitude)
	// visible, distance, elevation := SatelliteVisible(&gs, satellite_data.Position[0], startTime)
	// log.Println(visible, distance, elevation)
	// Test passed
}
