package space

import (
	"bufio"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	gosat "github.com/joshuaferrara/go-satellite"
)

type LatLong struct {
	Latitude  float64 `parquet:"lattitude"`
	Longitude float64 `parquet:"longitude"`
}

func (ll LatLong) asGosatLatLone() gosat.LatLong {
	return gosat.LatLong{
		Latitude:  ll.Latitude,
		Longitude: ll.Longitude,
	}
}

type GroundStation struct {
	Title    string `parquet:"groundstation_title"`
	ID       int    `parquet:"groundstation_id"`
	Latlong  LatLong
	Position []Vector3
	IsAP     bool `parquet:"is_access_point"`
	Isactive bool `parquet:"is_active"`
	// altitude float64
}

// Calculate distance from point A(Groundstation) to point B(satellite) on a sphere using Latitude and Longitude.
func CalculateDistance(latlong1 LatLong, latlong2 LatLong, max_distance float64) (d float64, inCircle bool) {
	R := 6378.0
	lat1_r := latlong1.Latitude * math.Pi / 180
	lat2_r := latlong2.Latitude * math.Pi / 180
	lat_d := (latlong1.Latitude - latlong2.Latitude) * math.Pi / 180
	long_d := (latlong1.Longitude - latlong2.Longitude) * math.Pi / 180

	//a := math.Sin(lat_d/2)*math.Sin(lat_d/2) + math.Cos(lat1_r) * math.Cos(lat2_r) * math.Sin(long_d/2) * math.Sin(long_d/2)
	//c := 2*math.Atan2(math.Sqrt(a),math.Sqrt(1-a))

	c := 2 * math.Asin(math.Sqrt(math.Sin(lat_d/2)*math.Sin(lat_d/2)+math.Cos(lat1_r)*math.Cos(lat2_r)*math.Sin(long_d/2)*math.Sin(long_d/2)))
	d = R * c
	if d > max_distance {
		inCircle = false
	} else {
		inCircle = true
	}

	return d, inCircle
}

// const minimumElevation = 25
const minimumElevation = 25
const footprint = 1500

func SatelliteVisible(gs *GroundStation, satellite_ll LatLong) (visible bool, distance float64) {
	// jday := gosat.JDay(time.Year(), int(time.Month()), time.Day(), time.Hour(), time.Minute(), time.Second())

	// lookangle := gosat.ECIToLookAngles(satellitePosition.AsgosatVector(), gs.Latlong.asGosatLatLone(), 10, jday)
	// gs.Latlong
	d, incircle := CalculateDistance(gs.Latlong, satellite_ll, footprint)
	// elevation = lookangle.El * (180 / math.Pi)
	// https://github.com/joshuaferrara/go-satellite/issues/13
	visible = incircle
	distance = d
	// visible = (elevation > minimumElevation) && elevation > 0
	// log.Info().Float64("elevation", elevation).Bool("visible", visible).Float64("range", lookangle.Rg).Msg("look angle")
	return visible, distance
}

func AccessPointVisible(gs1, gs2 *GroundStation, maxDistance float64) (inRange bool, distance float64) {
	distance, inRange = CalculateDistance(gs1.Latlong, gs2.Latlong, maxDistance)

	return inRange, distance
}

func GroundStationGen(title string, id int, lat float64, long float64, isAccessPoint bool) (GS GroundStation) {
	return GroundStation{
		Title: title,
		ID:    id,
		Latlong: LatLong{
			Latitude:  lat,
			Longitude: long,
		},
		IsAP: isAccessPoint,
	}
}

// func GroundstationPositionCal(GS1 groundStation, jday float64, altitude float64) gosat.Vector3 {
// 	GS1.position = gosat.LLAToECI(GS1.latlong, altitude, jday)

// 	return GS1.position
// }

func LoadGroundStations(path string) ([]GroundStation, error) {
	var groundStations []GroundStation
	f, err := os.Open(path)
	if err != nil {
		return groundStations, errors.New("could not open file")
	}
	defer f.Close()

	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)
	for GroundStationID := 0; fileScanner.Scan(); GroundStationID++ { // returns false at end of file
		line := fileScanner.Text()
		line_elements := strings.Split(line, ",")
		if len(line_elements) != 4 {
			return groundStations, errors.New("failed parsing line")
		}

		title := strings.TrimSpace(line_elements[0])
		lattiude, err := strconv.ParseFloat(strings.TrimSpace(line_elements[1]), 32)
		if err != nil {
			return groundStations, err
		}

		longitude, err := strconv.ParseFloat(strings.TrimSpace(line_elements[2]), 32)
		if err != nil {
			return groundStations, err
		}

		isAP, err := strconv.ParseBool(strings.TrimSpace(line_elements[3]))
		if err != nil {
			return groundStations, err
		}
		gs := GroundStationGen(title, GroundStationID, lattiude, longitude, isAP)
		groundStations = append(groundStations, gs)
	}
	return groundStations, nil
}

func GroundStationECIPostions(gs GroundStation, startTime time.Time, timestep time.Duration, duration time.Duration) []Vector3 {
	var altitude float64 = 0
	var endTime time.Time = startTime // don't touch this
	endTime = endTime.Add(duration)
	var positions []Vector3
	for i := 0; startTime.Before(endTime); i++ {
		jday := gosat.JDay(startTime.Year(), int(startTime.Month()), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second())
		position := gosat.LLAToECI(gs.Latlong.asGosatLatLone(), float64(altitude), jday)
		startTime = startTime.Add(timestep)
		newPos := Vector3{
			X: position.X,
			Y: position.Y,
			Z: position.Z,
		}
		positions = append(positions, newPos)
	}
	return positions
}
