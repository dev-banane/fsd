package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sampleWhazzup = `![DateStamp]17/08/2026 16:20:00
!GENERAL
VERSION = 1
CONNECTED CLIENTS = 1
!CLIENTS
DLH123:1000001:Test Pilot:PILOT::50.030000:8.570000:3000:250:B738:400:EDDF:FL350:EDDM:LAB:9:1:2000:0:40:0:I:1200:1210:1:30:2:0:EDDN:rmk:route::::::17/08/2026 16:20:00:0
!SERVERS
LAB:localhost:Lab:Lab FSD:1
`

func writeWhazzup(t *testing.T, age time.Duration) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "whazzup.txt")
	if err := os.WriteFile(path, []byte(sampleWhazzup), 0o644); err != nil {
		t.Fatalf("write whazzup: %v", err)
	}
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	return path
}

func TestRefreshFreshFileIsNotStale(t *testing.T) {
	c := NewCache(writeWhazzup(t, 0), 60, 10*time.Minute)
	if err := c.Refresh(); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got := c.Snapshot(); got.Stale {
		t.Fatal("fresh whazzup reported as stale")
	}
}

func TestRefreshFrozenFileGoesStale(t *testing.T) {
	c := NewCache(writeWhazzup(t, 0), 60, 10*time.Minute)
	if err := c.Refresh(); err != nil {
		t.Fatalf("initial refresh: %v", err)
	}
	if len(c.Snapshot().Pilots) != 1 {
		t.Fatalf("expected the sample pilot to parse, got %d", len(c.Snapshot().Pilots))
	}

	when := time.Now().Add(-staleAfter - time.Minute)
	if err := os.Chtimes(c.path, when, when); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	c.modTime = when

	if err := c.Refresh(); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if !c.Snapshot().Stale {
		t.Fatal("frozen whazzup still reported as live")
	}
}

func TestRefreshMissingFileIsStale(t *testing.T) {
	c := NewCache(filepath.Join(t.TempDir(), "gone.txt"), 60, 10*time.Minute)
	if err := c.Refresh(); err == nil {
		t.Fatal("expected an error for a missing whazzup file")
	}
	if !c.Snapshot().Stale {
		t.Fatal("missing whazzup not reported as stale")
	}
}
