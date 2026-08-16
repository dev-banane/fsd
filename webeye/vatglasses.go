package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed data/vatglasses
var vatglassesData embed.FS

type vgPosition struct {
	Pre       []string `json:"pre"`
	Type      string   `json:"type"`
	Frequency string   `json:"frequency"`
	Callsign  string   `json:"callsign"`
	Colours   []struct {
		Hex string `json:"hex"`
	} `json:"colours"`
}

type vgSector struct {
	Min    int        `json:"min"`
	Max    int        `json:"max"`
	Points [][]string `json:"points"`
}

type vgAirspace struct {
	ID      string     `json:"id"`
	Group   string     `json:"group"`
	UID     string     `json:"uid"`
	Owner   []string   `json:"owner"`
	Sectors []vgSector `json:"sectors"`
}

type vgFile struct {
	Airspace  []vgAirspace          `json:"airspace"`
	Positions map[string]vgPosition `json:"positions"`
}

type vgDataset struct {
	airspace  []vgAirspace
	positions map[string]vgPosition
}

var vatglasses = vgDataset{positions: map[string]vgPosition{}}

type Sector struct {
	Callsign string       `json:"callsign"`
	Position string       `json:"position"`
	Name     string       `json:"name"`
	Color    string       `json:"color"`
	Min      int          `json:"min"`
	Max      int          `json:"max"`
	Polygon  [][2]float64 `json:"polygon"`
}

func (d *vgDataset) merge(raw []byte, source string) {
	var file vgFile
	if err := json.Unmarshal(raw, &file); err != nil {
		log.Printf("webeye: vatglasses %s: %v", source, err)
		return
	}
	d.airspace = append(d.airspace, file.Airspace...)
	for uid, pos := range file.Positions {
		d.positions[uid] = pos
	}
}

func loadVatglasses() {
	vatglasses = vgDataset{positions: map[string]vgPosition{}}

	entries, err := fs.ReadDir(vatglassesData, "data/vatglasses")
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			raw, err := vatglassesData.ReadFile("data/vatglasses/" + e.Name())
			if err != nil {
				continue
			}
			vatglasses.merge(raw, e.Name())
		}
	}

	dir := os.Getenv("WEBEYE_VATGLASSES")
	if dir == "" {
		log.Printf("webeye: vatglasses: %d volumes, %d positions (bundled)",
			len(vatglasses.airspace), len(vatglasses.positions))
		return
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, path := range matches {
		raw, err := os.ReadFile(path)
		if err != nil {
			log.Printf("webeye: vatglasses %s: %v", path, err)
			continue
		}
		vatglasses.merge(raw, path)
	}
	log.Printf("webeye: vatglasses: %d volumes, %d positions (bundled + %s)",
		len(vatglasses.airspace), len(vatglasses.positions), dir)
}

func parseVgCoord(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	sign := 1.0
	switch s[0] {
	case '-':
		sign, s = -1, s[1:]
	case '+':
		s = s[1:]
	}
	if len(s) < 5 {
		return 0, false
	}

	deg, err1 := strconv.Atoi(s[:len(s)-4])
	min, err2 := strconv.Atoi(s[len(s)-4 : len(s)-2])
	sec, err3 := strconv.Atoi(s[len(s)-2:])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, false
	}
	return sign * (float64(deg) + float64(min)/60 + float64(sec)/3600), true
}

func splitCallsign(callsign string) (prefix, uid, suffix string) {
	parts := strings.Split(strings.ToUpper(callsign), "_")
	if len(parts) < 2 {
		return "", "", ""
	}
	prefix = parts[0]
	suffix = parts[len(parts)-1]
	if len(parts) >= 3 {
		uid = parts[len(parts)-2]
	}
	return
}

func sectorTypeOf(suffix string) string {
	switch suffix {
	case "CTR", "FSS":
		return "CTR"
	case "APP", "DEP":
		return "APP"
	}
	return ""
}

func typeMatches(posType, want string) bool {
	posType = strings.ToUpper(posType)
	if want == "APP" {
		return posType == "APP" || posType == "DEP"
	}
	return posType == want
}

func ownedPositions(callsign string) []string {
	prefix, uid, suffix := splitCallsign(callsign)
	want := sectorTypeOf(suffix)
	if prefix == "" || want == "" {
		return nil
	}

	hasPrefix := func(p vgPosition) bool {
		for _, pre := range p.Pre {
			if strings.EqualFold(pre, prefix) {
				return true
			}
		}
		return false
	}

	if uid != "" {
		if pos, ok := vatglasses.positions[uid]; ok &&
			hasPrefix(pos) && typeMatches(pos.Type, want) {
			return []string{uid}
		}
	}

	var owned []string
	for id, pos := range vatglasses.positions {
		if hasPrefix(pos) && typeMatches(pos.Type, want) {
			owned = append(owned, id)
		}
	}
	return owned
}

const unlimitedCeiling = 999

func ceilingOf(max int) int {
	if max <= 0 {
		return unlimitedCeiling
	}
	return max
}

func resolveSectors(controllers []Controller) []Sector {
	if len(vatglasses.airspace) == 0 {
		return []Sector{}
	}

	held := map[string]string{}
	meta := map[string]Controller{}
	for _, c := range controllers {
		if sectorTypeOf(strings.ToUpper(c.Position)) == "" {
			continue
		}
		for _, uid := range ownedPositions(c.Callsign) {
			if _, taken := held[uid]; !taken {
				held[uid] = c.Callsign
			}
		}
		meta[c.Callsign] = c
	}
	if len(held) == 0 {
		return []Sector{}
	}

	sectors := []Sector{}
	for _, volume := range vatglasses.airspace {
		callsign, uid := "", ""
		for _, candidate := range volume.Owner {
			if cs, ok := held[candidate]; ok {
				callsign, uid = cs, candidate
				break
			}
		}
		if callsign == "" {
			continue
		}

		controller := meta[callsign]
		pos := vatglasses.positions[uid]
		colour := ""
		if len(pos.Colours) > 0 {
			colour = pos.Colours[0].Hex
		}

		for _, s := range volume.Sectors {
			polygon := make([][2]float64, 0, len(s.Points))
			var prev [2]float64
			for _, p := range s.Points {
				if len(p) != 2 {
					continue
				}
				lat, ok1 := parseVgCoord(p[0])
				lon, ok2 := parseVgCoord(p[1])
				if !ok1 || !ok2 {
					continue
				}
				point := [2]float64{lat, lon}
				if len(polygon) > 0 && point == prev {
					continue
				}
				polygon = append(polygon, point)
				prev = point
			}
			if len(polygon) < 3 {
				continue
			}
			sectors = append(sectors, Sector{
				Callsign: callsign,
				Position: controller.Position,
				Name:     pos.Callsign,
				Color:    colour,
				Min:      s.Min,
				Max:      ceilingOf(s.Max),
				Polygon:  polygon,
			})
		}
	}
	return sectors
}
