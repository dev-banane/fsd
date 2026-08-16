package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testStatic() fstest.MapFS {
	return fstest.MapFS{
		"index.html":        &fstest.MapFile{Data: []byte("<html>shell</html>")},
		"assets/app.js":     &fstest.MapFile{Data: []byte("console.log(1)")},
		"aircraft/a320.svg": &fstest.MapFile{Data: []byte("<svg></svg>")},
		"assets/worker.mjs": &fstest.MapFile{Data: []byte("export {}")},
	}
}

func TestSpaHandlerMjsContentType(t *testing.T) {
	h := spaHandler(testStatic(), "WebEye")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/worker.mjs", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /assets/worker.mjs = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("Content-Type = %q, want text/javascript", ct)
	}
}

func TestSpaHandlerMissingAssetIs404(t *testing.T) {
	h := spaHandler(testStatic(), "WebEye")
	for _, path := range []string{"/assets/missing.js", "/aircraft/missing.svg", "/does-not-exist.png"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
	}
}

func TestSpaHandlerRouteFallsBackToShell(t *testing.T) {
	h := spaHandler(testStatic(), "WebEye")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/some/deep/route", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /some/deep/route = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "<html>shell</html>" {
		t.Errorf("body = %q, want the index.html shell", rec.Body.String())
	}
}

func TestSpaHandlerServesRealAssets(t *testing.T) {
	h := spaHandler(testStatic(), "WebEye")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /assets/app.js = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q, want immutable for /assets/", got)
	}
}

func TestSpaHandlerInjectsTitle(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(
			`<html><head><title>WebEye</title><meta name="webeye-title" content="WebEye"></head></html>`,
		)},
	}
	h := spaHandler(fsys, `Night & "Watch"`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "<title>Night &amp; &#34;Watch&#34;</title>") {
		t.Errorf("title tag = %q", body)
	}
	if !strings.Contains(body, `content="Night &amp; &#34;Watch&#34;"`) {
		t.Errorf("title meta = %q", body)
	}
}

func TestLoadConfigTitle(t *testing.T) {
	t.Setenv("WEBEYE_TITLE", "EDGG Radar")
	if got := loadConfig().title; got != "EDGG Radar" {
		t.Errorf("title = %q, want EDGG Radar", got)
	}
}
