package space

import (
	"math"
	"runtime"
	"sort"
	"time"

	gosat "github.com/joshuaferrara/go-satellite"
)

type Vector3 struct {
	X float64 `parquet:"X"`
	Y float64 `parquet:"Y"`
	Z float64 `parquet:"Z"`
	// Vector gosat.Vector3
}

const (
	r          float64 = 6378 // Radius of earth
	atmosphere float64 = 80   // Mesosphere
)

func (vector Vector3) AsgosatVector() gosat.Vector3 {
	return gosat.Vector3{
		X: vector.X,
		Y: vector.Y,
		Z: vector.Z,
	}
}

func (vectorA Vector3) Sub(vectorB Vector3) (vectorAB Vector3) {
	vectorAB.X = vectorA.X - vectorB.X
	vectorAB.Y = vectorA.Y - vectorB.Y
	vectorAB.Z = vectorA.Z - vectorB.Z
	return vectorAB
}

func (vectorA Vector3) Speed() float64 {
	return vectorA.magnitude()
}

func (vectorA Vector3) DopplerShift(vectorB Vector3) float64 {
	v_o := vectorA.Speed()
	v_s := vectorA.Sub(vectorB).Speed()
	f_s := 430e12
	v := C
	f_o := ((v + v_o) / (v + v_s)) * f_s
	return math.Abs(f_o - f_s)
}

func (vectorA Vector3) magnitude() float64 {
	return math.Sqrt(vectorA.X*vectorA.X + vectorA.Y*vectorA.Y + vectorA.Z + vectorA.Z)
}

// h=350
func LineOfSight(satellite_height float64) float64 {
	return math.Sqrt(math.Pow((r+satellite_height), 2)-math.Pow((r+atmosphere), 2)) * 2.0
}

func (p1 Vector3) Distance(p2 Vector3) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2) + math.Pow(p1.Z-p2.Z, 2))
}

func Reachable(p1 Vector3, p2 Vector3, linkDistance float64) bool {
	return p1.Distance(p2) < linkDistance
}

func DistanceVector(satellite1, satellite2 []Vector3) (distance []float64) {
	distance = make([]float64, len(satellite1))
	for i := 0; i < len(satellite1); i++ {
		distance[i] = satellite1[i].Distance(satellite2[i])
	}
	return distance
}

func newVector(x, y, z float64) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

const C = 299792.458 // km/s

func Latency(distance float64) float64 {
	return distance / C
}

func LatencyVector(distance []float64) (latency []float64) {
	latency = make([]float64, len(distance))
	for i := 0; i < len(distance); i++ {
		latency[i] = Latency(distance[i])
	}
	return latency
}

type OrbitalData struct {
	Isactive    bool   `parquet:"is_active"`
	SatelliteId int    `parquet:"satellite_id"`
	Title       string `parquet:"satellite_title"`
	Position    []Vector3
	Velocity    []Vector3
	LatLong     []LatLong
	//Time_steps []time.Time
}

func getSatPos(sat_channel <-chan satellite, satData chan<- OrbitalData, startTime time.Time, timestep time.Duration, duration time.Duration) {
	for sat := range sat_channel {
		var data OrbitalData
		data.Position = make([]Vector3, duration/timestep)
		data.Velocity = make([]Vector3, duration/timestep)
		data.LatLong = make([]LatLong, duration/timestep)
		localStartTime := startTime
		for i := 0; i < int(duration)/int(timestep); i++ {
			localStartTime = localStartTime.Add(1 * timestep)
			// gst := gosat.GSTimeFromDate(startTime.Year(), int(startTime.Month()), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second())
			pos, vel := gosat.Propagate(sat.gosat, localStartTime.Year(), int(localStartTime.Month()), localStartTime.Day(), localStartTime.Hour(), localStartTime.Minute(), localStartTime.Second())

			// jday := gosat.JDay(localStartTime.Year(), int(localStartTime.Month()), localStartTime.Day(), localStartTime.Hour(), localStartTime.Minute(), localStartTime.Second())
			// gmst := gosat.ThetaG_JD(jday)

			// convert the current time to Galileo system time (GST)
			gst := gosat.GSTimeFromDate(localStartTime.Year(), int(localStartTime.Month()), localStartTime.Day(), localStartTime.Hour(), localStartTime.Minute(), localStartTime.Second())

			_, _, ll := gosat.ECIToLLA(pos, gst)

			ll_deg := gosat.LatLongDeg(ll)
			// altitude, velocity, ll := gosat.ECIToLLA(pos, gst)
			//fmt.Println(pos)
			data.Position[i] = Vector3{X: pos.X, Y: pos.Y, Z: pos.Z}
			data.Velocity[i] = Vector3{X: vel.X, Y: vel.Y, Z: vel.Z}
			data.LatLong[i] = LatLong{
				Latitude:  ll_deg.Latitude,
				Longitude: ll_deg.Longitude,
			}
		}
		data.SatelliteId = sat.satelliteId
		satData <- data
	}
}

func LLAFromPosition(pos Vector3, time time.Time) LatLong {
	gst := gosat.GSTimeFromDate(time.Year(), int(time.Month()), time.Day(), time.Hour(), time.Minute(), time.Second())
	//al, vel, ll
	_, _, ll := gosat.ECIToLLA(pos.AsgosatVector(), gst)
	ll_deg := gosat.LatLongDeg(ll)
	return LatLong{Latitude: ll_deg.Latitude, Longitude: ll_deg.Longitude}
}

// Sort interface implementation
type OrbitalDataByID []OrbitalData

func (satellites OrbitalDataByID) Len() int {
	return len(satellites)
}

func (satellites OrbitalDataByID) Swap(i, j int) {
	satellites[i], satellites[j] = satellites[j], satellites[i]
}

func (satellites OrbitalDataByID) Less(i, j int) bool {
	return satellites[i].SatelliteId < satellites[j].SatelliteId
}

type satellite struct {
	satelliteId int `parquet:"satellite_id"`
	gosat       gosat.Satellite
}

func GetSatData(satellites []gosat.Satellite, satelliteids []int, startTime time.Time, timestep time.Duration, duration time.Duration) (orbitalData []OrbitalData) {

	jobcount := len(satellites)
	jobs := make(chan satellite, jobcount)
	results := make(chan OrbitalData, jobcount)
	// jobs <-chan int, results chan<- int
	for w := 1; w <= runtime.NumCPU(); w++ {
		go getSatPos(jobs, results, startTime, timestep, duration)
	}
	for i, gosat := range satellites {
		temporary_sat := satellite{
			satelliteId: satelliteids[i],
			gosat:       gosat,
		}
		jobs <- temporary_sat
	}
	for i := 0; i < jobcount; i++ {
		satellitePositions := <-results
		orbitalData = append(orbitalData, satellitePositions)
	}
	close(jobs)

	// TODO sort (Mulvad says this works now) please
	sort.Sort(OrbitalDataByID(orbitalData))

	return orbitalData
}
