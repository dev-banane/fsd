package main

import (
	"bufio"
	"embed"
	"strconv"
	"strings"
)

//go:embed data/airports.csv
var airportData embed.FS

var airports = map[string]LatLon{}

func loadAirports() error {
	fh, err := airportData.Open("data/airports.csv")
	if err != nil {
		return err
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",")
		if len(parts) != 3 {
			continue
		}
		lat, err1 := strconv.ParseFloat(parts[1], 64)
		lon, err2 := strconv.ParseFloat(parts[2], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		airports[parts[0]] = LatLon{Lat: lat, Lon: lon}
	}
	return scanner.Err()
}

func lookupAirport(icao string) *LatLon {
	if len(icao) != 4 {
		return nil
	}
	if pos, ok := airports[strings.ToUpper(icao)]; ok {
		return &pos
	}
	return nil
}

func resolveRoute(p *Pilot) {
	p.DepPos = lookupAirport(p.Departure)
	p.ArrPos = lookupAirport(p.Arrival)
}

func resolveStation(c *Controller) {
	if c.Airport == "" {
		return
	}
	c.AirportPos = lookupAirport(c.Airport)
}
