package main

import (
	"os"
	"sync"
	"time"
)

type Status struct {
	Timestamp   string                  `json:"timestamp"`
	ClientCount int                     `json:"clientCount"`
	Controllers []Controller            `json:"controllers"`
	Pilots      []Pilot                 `json:"pilots"`
	Servers     []ServerInfo            `json:"servers"`
	Sectors     []Sector                `json:"sectors"`
	History     map[string][][2]float64 `json:"history"`
	Stale       bool                    `json:"stale"`
}

func emptyStatus() *Status {
	return &Status{
		Controllers: []Controller{},
		Pilots:      []Pilot{},
		Servers:     []ServerInfo{},
		Sectors:     []Sector{},
		History:     map[string][][2]float64{},
		Stale:       true,
	}
}

const staleAfter = 90 * time.Second

type Cache struct {
	path          string
	historyPoints int
	historyTTL    time.Duration

	mu      sync.RWMutex
	payload *Status

	modTime  time.Time
	history  map[string][][2]float64
	lastSeen map[string]time.Time
}

func NewCache(path string, historyPoints int, historyTTL time.Duration) *Cache {
	return &Cache{
		path:          path,
		historyPoints: historyPoints,
		historyTTL:    historyTTL,
		payload:       emptyStatus(),
		history:       map[string][][2]float64{},
		lastSeen:      map[string]time.Time{},
	}
}

func (c *Cache) Snapshot() *Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.payload
}

func (c *Cache) markStale() {
	c.mu.Lock()
	defer c.mu.Unlock()
	stale := *c.payload
	stale.Stale = true
	c.payload = &stale
}

func (c *Cache) Refresh() error {
	info, err := os.Stat(c.path)
	if err != nil {
		c.markStale()
		return err
	}
	if info.ModTime().Equal(c.modTime) {
		if time.Since(info.ModTime()) > staleAfter {
			c.markStale()
		}
		return nil
	}
	c.modTime = info.ModTime()

	raw, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}

	controllers, pilots, servers, meta := ParseWhazzup(string(raw))
	for i := range pilots {
		resolveRoute(&pilots[i])
	}
	for i := range controllers {
		resolveStation(&controllers[i])
	}
	c.recordHistory(pilots)

	count := atoiSafe(meta["CONNECTED CLIENTS"])
	if count == 0 {
		count = len(controllers) + len(pilots)
	}

	history := make(map[string][][2]float64, len(c.history))
	for callsign, track := range c.history {
		clone := make([][2]float64, len(track))
		copy(clone, track)
		history[callsign] = clone
	}

	payload := &Status{
		Timestamp:   info.ModTime().UTC().Format("02/01/2006 15:04:05"),
		ClientCount: count,
		Controllers: controllers,
		Pilots:      pilots,
		Servers:     servers,
		Sectors:     resolveSectors(controllers),
		History:     history,
		Stale:       false,
	}

	c.mu.Lock()
	c.payload = payload
	c.mu.Unlock()
	return nil
}

func (c *Cache) recordHistory(pilots []Pilot) {
	now := time.Now()
	for _, p := range pilots {
		if p.Lat == nil || p.Lon == nil || (*p.Lat == 0 && *p.Lon == 0) {
			continue
		}
		point := [2]float64{roundCoord(*p.Lat), roundCoord(*p.Lon)}
		track := c.history[p.Callsign]
		if n := len(track); n == 0 || track[n-1] != point {
			track = append(track, point)
			if len(track) > c.historyPoints {
				track = track[len(track)-c.historyPoints:]
			}
			c.history[p.Callsign] = track
		}
		c.lastSeen[p.Callsign] = now
	}

	for callsign, seen := range c.lastSeen {
		if now.Sub(seen) > c.historyTTL {
			delete(c.lastSeen, callsign)
			delete(c.history, callsign)
		}
	}
}

func roundCoord(f float64) float64 {
	return float64(int(f*100000+copysign(0.5, f))) / 100000
}
