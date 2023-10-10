package tle

import (
	"bufio"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	gosat "github.com/joshuaferrara/go-satellite"
	"github.com/rs/zerolog/log"
)

type TLEdata struct {
	Title string `parquet:"title_line"`
	Line1 string `parquet:"l1"`
	Line2 string `parquet:"l2"`
}

func loadTLEdata(filename string) (satellites []TLEdata, found bool) {
	var lines []string
	f, err := os.Open(filename)
	if err != nil {
		return satellites, false
	}

	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)

	for i := 0; fileScanner.Scan(); i++ {
		lines = append(lines, fileScanner.Text())
	}
	defer f.Close()

	if len(lines)%3 != 0 {
		log.Warn().Msg("Lines should be a multiple of 3")
	}

	for i := 0; i < len(lines)-2; i += 3 {
		satellites = append(satellites, TLEdata{
			Title: lines[i],
			Line1: lines[i+1],
			Line2: lines[i+2],
		})
	}
	return satellites, true
}

func idFromTitleLine(title string) int {
	var id_string string = title
	if strings.HasPrefix(title, "STARLINK-") {
		id_string = strings.TrimPrefix(title, "STARLINK-")
	}
	if strings.HasPrefix(title, "ONEWEB-") {
		id_string = strings.TrimPrefix(title, "ONEWEB-")
	}

	id_string = strings.TrimSpace(id_string)
	id_string = strings.TrimSuffix(id_string, " (DARKSAT)")
	id_string = strings.TrimSuffix(id_string, " (VISORSAT)")

	id_string = strings.TrimSpace(id_string)
	id, err := strconv.Atoi(id_string)
	if err != nil {
		log.Warn().Msgf("Could not parse id from title line: %s", title)
		return rand.Intn(3000)
	}
	return id
}

func LoadSatellites(filename string) (satelliteIds []int, satellites []gosat.Satellite, found bool) {
	tledata, found := loadTLEdata(filename)
	if !found {
		return nil, nil, false
	}
	for _, tle := range tledata {
		satellite := gosat.TLEToSat(tle.Line1, tle.Line2, "wgs84")
		satellites = append(satellites, satellite)
		id := idFromTitleLine(tle.Title)
		satelliteIds = append(satelliteIds, id)
		// gosat.ParseTLE(tle.Line1, tle.Line2, "wgs84")
	}
	return satelliteIds, satellites, true
}

// lattitudes []float64
// longitudes []float64

func GetTLEfromNoradIDs(spacetrack *gosat.Spacetrack, norad_ids []int) (satellites []gosat.Satellite) {
	for _, sattellite_id := range norad_ids {
		sat, err := spacetrack.GetTLE(uint64(sattellite_id), time.Now(), "wgs84")
		// Todo fixed time to make reproducible
		if err != nil {
			log.Fatal().Err(err).Msg("DIE")
		}
		// log.Println(sat.Line1, sat.Line2)
		satellites = append(satellites, sat)
	}
	return satellites
}
