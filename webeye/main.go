package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func init() {
	if err := mime.AddExtensionType(".mjs", "text/javascript; charset=utf-8"); err != nil {
		log.Printf("webeye: could not register .mjs MIME type: %v", err)
	}
}

//go:embed all:static
var staticFiles embed.FS

type config struct {
	whazzup       string
	addr          string
	poll          time.Duration
	historyPoints int
	historyTTL    time.Duration
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envSeconds(key string, fallback float64) time.Duration {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Duration(f * float64(time.Second))
		}
	}
	return time.Duration(fallback * float64(time.Second))
}

func loadConfig() config {
	return config{
		whazzup:       env("WEBEYE_WHAZZUP", "/data/whazzup.txt"),
		addr:          ":" + env("WEBEYE_PORT", "8080"),
		poll:          envSeconds("WEBEYE_POLL", 5),
		historyPoints: envInt("WEBEYE_HISTORY_POINTS", 60),
		historyTTL:    envSeconds("WEBEYE_HISTORY_TTL", 600),
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("webeye: encode failed: %v", err)
	}
}

func spaHandler(root fs.FS) http.Handler {
	files := http.FS(root)
	server := http.FileServer(files)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean("/" + r.URL.Path)

		if f, err := files.Open(clean); err == nil {
			info, statErr := f.Stat()
			f.Close()
			if statErr == nil && !info.IsDir() {
				if strings.HasPrefix(clean, "/assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "no-cache")
				}
				server.ServeHTTP(w, r)
				return
			}
		}

		if strings.Contains(path.Base(clean), ".") {
			http.NotFound(w, r)
			return
		}

		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		w.Header().Set("Cache-Control", "no-cache")
		server.ServeHTTP(w, r2)
	})
}

func main() {
	cfg := loadConfig()

	if err := loadAirports(); err != nil {
		log.Printf("webeye: airport table unavailable, routes disabled: %v", err)
	}
	loadVatglasses()

	cache := NewCache(cfg.whazzup, cfg.historyPoints, cfg.historyTTL)
	if err := cache.Refresh(); err != nil {
		log.Printf("webeye: initial read of %s failed: %v", cfg.whazzup, err)
	}

	go func() {
		ticker := time.NewTicker(cfg.poll)
		defer ticker.Stop()
		for range ticker.C {
			if err := cache.Refresh(); err != nil {
				log.Printf("webeye: poll failed: %v", err)
			}
		}
	}()

	assets, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("webeye: embedded frontend missing: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/fsd-status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, cache.Snapshot())
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]bool{"ok": !cache.Snapshot().Stale})
	})
	mux.Handle("/", spaHandler(assets))

	server := &http.Server{
		Addr:              cfg.addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("webeye: serving on %s, reading %s every %s",
		cfg.addr, cfg.whazzup, cfg.poll)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("webeye: %v", err)
	}
}
