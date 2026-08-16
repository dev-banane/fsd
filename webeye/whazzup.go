package main

import (
	"strconv"
	"strings"
)

const (
	fCallsign = 0
	fCID      = 1
	fRealName = 2
	fType     = 3

	fFrequency = 4

	fLat         = 5
	fLon         = 6
	fAltitude    = 7
	fGroundspeed = 8

	fAircraft    = 9
	fTASCruise   = 10
	fDepAirport  = 11
	fPlanAlt     = 12
	fDestAirport = 13

	fServer      = 14
	fProtocol    = 15
	fRating      = 16
	fTransponder = 17
	fFacility    = 18
	fVisRange    = 19

	fFlightRules = 21

	minFields = 22
)

type Controller struct {
	Callsign   string   `json:"callsign"`
	CID        string   `json:"cid"`
	Name       string   `json:"name"`
	Lat        *float64 `json:"lat"`
	Lon        *float64 `json:"lon"`
	Server     string   `json:"server"`
	Rating     int      `json:"rating"`
	Logon      string   `json:"logon"`
	Frequency  string   `json:"frequency"`
	Range      int      `json:"range"`
	Facility   int      `json:"facility"`
	Position   string   `json:"position"`
	Airport    string   `json:"airport"`
	AirportPos *LatLon  `json:"airport_pos,omitempty"`
}

type Pilot struct {
	Callsign    string   `json:"callsign"`
	CID         string   `json:"cid"`
	Name        string   `json:"name"`
	Lat         *float64 `json:"lat"`
	Lon         *float64 `json:"lon"`
	Server      string   `json:"server"`
	Rating      int      `json:"rating"`
	Logon       string   `json:"logon"`
	Alt         int      `json:"alt"`
	Groundspeed int      `json:"groundspeed"`
	Type        string   `json:"type"`
	Departure   string   `json:"departure_ap"`
	Arrival     string   `json:"arrival_ap"`
	Squawk      string   `json:"sqwk"`
	CruiseLevel int      `json:"crzlvl"`
	TAS         int      `json:"tas"`
	FlightRules string   `json:"flight_rules"`
	Heading     float64  `json:"heading"`
	Pitch       float64  `json:"pitch"`
	Bank        float64  `json:"bank"`
	OnGround    bool     `json:"onground"`
	DepPos      *LatLon  `json:"departure_pos,omitempty"`
	ArrPos      *LatLon  `json:"arrival_pos,omitempty"`
}

type LatLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type ServerInfo struct {
	Ident    string `json:"ident"`
	Hostname string `json:"hostname"`
	Location string `json:"location"`
	Name     string `json:"name"`
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

func atofSafe(s string) *float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil
	}
	return &f
}

// PBH unpacks FSD's 32-bit pitch/bank/heading word: 10 bits pitch, 10 bits
// bank, 10 bits heading, 1 bit on-ground, 1 bit unused. Each angle is in
// units of 360/1024 degrees.
func decodePBH(raw string) (pitch, bank, heading float64, onGround bool) {
	v, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, 0, 0, false
	}
	pbh := uint32(v)
	toDeg := func(x uint32) float64 { return float64(x) * 360.0 / 1024.0 }
	signed := func(d float64) float64 {
		if d > 180 {
			return d - 360
		}
		return d
	}
	pitch = round1(signed(toDeg((pbh >> 22) & 0x3FF)))
	bank = round1(signed(toDeg((pbh >> 12) & 0x3FF)))
	heading = round1(toDeg((pbh >> 2) & 0x3FF))
	onGround = (pbh>>1)&1 == 1
	return
}

func round1(f float64) float64 {
	return float64(int(f*10+copysign(0.5, f))) / 10
}

func copysign(mag, sign float64) float64 {
	if sign < 0 {
		return -mag
	}
	return mag
}

func positionOf(callsign string) string {
	parts := strings.Split(strings.ToUpper(callsign), "_")
	switch suffix := parts[len(parts)-1]; suffix {
	case "DEL", "GND", "TWR", "ATIS":
		return suffix
	case "RMP":
		return "GND"
	case "APP", "DEP":
		return "APP"
	case "CTR", "FSS":
		return "CTR"
	default:
		return "OBS"
	}
}

func airportOf(callsign string) string {
	prefix, _, found := strings.Cut(strings.ToUpper(callsign), "_")
	if !found || len(prefix) < 3 || len(prefix) > 4 {
		return ""
	}
	for _, r := range prefix {
		if r < 'A' || r > 'Z' {
			return ""
		}
	}
	return prefix
}

func ParseWhazzup(text string) (controllers []Controller, pilots []Pilot, servers []ServerInfo, meta map[string]string) {
	controllers, pilots, servers = []Controller{}, []Pilot{}, []ServerInfo{}
	meta = map[string]string{}
	section := ""

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(strings.TrimSpace(raw), "\r")
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "![") {
			continue
		}
		if strings.HasPrefix(line, "!") {
			section = strings.ToUpper(strings.TrimPrefix(line, "!"))
			continue
		}

		switch section {
		case "GENERAL":
			if key, value, ok := strings.Cut(line, "="); ok {
				meta[strings.ToUpper(strings.TrimSpace(key))] = strings.TrimSpace(value)
			}
		case "CLIENTS":
			f := strings.Split(line, ":")
			if len(f) < minFields || f[fCallsign] == "" {
				continue
			}
			if strings.EqualFold(f[fType], "ATC") {
				controllers = append(controllers, parseController(f))
			} else {
				pilots = append(pilots, parsePilot(f))
			}
		case "SERVERS":
			f := strings.Split(line, ":")
			if len(f) >= 4 {
				servers = append(servers, ServerInfo{
					Ident: f[0], Hostname: f[1], Location: f[2], Name: f[3],
				})
			}
		}
	}
	return
}

func tail(f []string) (logon, pbh string) {
	if len(f) <= minFields {
		return "", ""
	}
	return f[len(f)-2], f[len(f)-1]
}

func parseController(f []string) Controller {
	logon, _ := tail(f)
	return Controller{
		Callsign:  f[fCallsign],
		CID:       f[fCID],
		Name:      f[fRealName],
		Lat:       atofSafe(f[fLat]),
		Lon:       atofSafe(f[fLon]),
		Server:    f[fServer],
		Rating:    atoiSafe(f[fRating]),
		Logon:     logon,
		Frequency: f[fFrequency],
		Range:     atoiSafe(f[fVisRange]),
		Facility:  atoiSafe(f[fFacility]),
		Position:  positionOf(f[fCallsign]),
		Airport:   airportOf(f[fCallsign]),
	}
}

func parsePilot(f []string) Pilot {
	logon, pbhRaw := tail(f)
	pitch, bank, heading, onGround := decodePBH(pbhRaw)

	rules := ""
	if f[fFlightRules] != "" {
		rules = strings.ToUpper(f[fFlightRules])[:1]
	}

	return Pilot{
		Callsign:    f[fCallsign],
		CID:         f[fCID],
		Name:        f[fRealName],
		Lat:         atofSafe(f[fLat]),
		Lon:         atofSafe(f[fLon]),
		Server:      f[fServer],
		Rating:      atoiSafe(f[fRating]),
		Logon:       logon,
		Alt:         atoiSafe(f[fAltitude]),
		Groundspeed: atoiSafe(f[fGroundspeed]),
		Type:        f[fAircraft],
		Departure:   strings.ToUpper(f[fDepAirport]),
		Arrival:     strings.ToUpper(f[fDestAirport]),
		Squawk:      f[fTransponder],
		CruiseLevel: atoiSafe(f[fPlanAlt]),
		TAS:         atoiSafe(f[fTASCruise]),
		FlightRules: rules,
		Heading:     heading,
		Pitch:       pitch,
		Bank:        bank,
		OnGround:    onGround,
	}
}
