package traefik_plugin_servefile_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	traefik_plugin_servefile "github.com/xuewenG/traefik-plugin-servefile"
)

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	must(t, os.WriteFile(p, []byte(content), 0644))
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "home.html"), "home.html")
	writeFile(t, filepath.Join(dir, "fallback.html"), "fallback.html")

	writeFile(t, filepath.Join(dir, "styles.css"), "")
	writeFile(t, filepath.Join(dir, "app.js"), "")
	writeFile(t, filepath.Join(dir, "photo.jpg"), "")
	writeFile(t, filepath.Join(dir, "photo.png"), "")
	writeFile(t, filepath.Join(dir, "photo.svg"), "")
	writeFile(t, filepath.Join(dir, "data.json"), "")
	writeFile(t, filepath.Join(dir, "data.xml"), "")
	writeFile(t, filepath.Join(dir, "video.mp4"), "")
	writeFile(t, filepath.Join(dir, "video.mkv"), "")
	writeFile(t, filepath.Join(dir, "archive.zip"), "")
	writeFile(t, filepath.Join(dir, "archive.tar"), "")
	writeFile(t, filepath.Join(dir, "archive.tar.gz"), "")
	writeFile(t, filepath.Join(dir, "archive.tgz"), "")
	writeFile(t, filepath.Join(dir, "raw.bin"), "\x00\x01\x02\x03\x04\x05")

	must(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0755))
	writeFile(t, filepath.Join(dir, "subdir", "home.html"), "subdir/home.html")

	must(t, os.MkdirAll(filepath.Join(dir, "emptydir"), 0755))

	return dir
}

func nopHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
}

func newHandler(t *testing.T, cfg *traefik_plugin_servefile.Config) http.Handler {
	t.Helper()
	h, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return h
}

func doRequest(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, httptest.NewRequest(method, path, nil))
	return rw
}

func assertStatus(t *testing.T, rw *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rw.Code != want {
		t.Errorf("status = %d, want %d", rw.Code, want)
	}
}

func assertBodyEqual(t *testing.T, rw *httptest.ResponseRecorder, want string) {
	t.Helper()
	if rw.Body.String() != want {
		t.Errorf("body = %q, want %q", rw.Body.String(), want)
	}
}

func assertContentType(t *testing.T, rw *httptest.ResponseRecorder, wantPrefix string) {
	t.Helper()
	if ct := rw.Header().Get("content-type"); !strings.HasPrefix(ct, wantPrefix) {
		t.Errorf("content-type = %q, want prefix %q", ct, wantPrefix)
	}
}

// ── TestNew ───────────────────────────────────────────────────────────────────

func TestNew(t *testing.T) {
	dir := setupTestDir(t)

	t.Run("error when rootDir is empty", func(t *testing.T) {
		cfg := traefik_plugin_servefile.CreateConfig()
		_, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error when rootDir does not exist", func(t *testing.T) {
		cfg := traefik_plugin_servefile.CreateConfig()
		cfg.RootDir = "/nonexistentdir"
		_, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error when rootDir is a file", func(t *testing.T) {
		cfg := traefik_plugin_servefile.CreateConfig()
		cfg.RootDir = filepath.Join(dir, "home.html")
		_, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success with minimal config", func(t *testing.T) {
		cfg := traefik_plugin_servefile.CreateConfig()
		cfg.RootDir = dir
		h, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})

	t.Run("success when IndexFile and FallbackFile do not exist", func(t *testing.T) {
		cfg := traefik_plugin_servefile.CreateConfig()
		cfg.RootDir = dir
		cfg.IndexFile = "nonexistent.html"
		cfg.FallbackFile = "nonexistent.html"
		h, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})

	t.Run("success when rootDir has a trailing slash", func(t *testing.T) {
		cfg := traefik_plugin_servefile.CreateConfig()
		cfg.RootDir = dir + "/"
		h, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})

	t.Run("success with all options", func(t *testing.T) {
		cfg := traefik_plugin_servefile.CreateConfig()
		cfg.RootDir = dir
		cfg.IndexFile = "home.html"
		cfg.FallbackFile = "fallback.html"
		cfg.ListDir = true
		h, err := traefik_plugin_servefile.New(context.Background(), nopHandler(), cfg, "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})
}

// ── TestServeHTTP ─────────────────────────────────────────────────────────────

func TestServeHTTP(t *testing.T) {
	base := func() *traefik_plugin_servefile.Config {
		cfg := traefik_plugin_servefile.CreateConfig()
		cfg.RootDir = setupTestDir(t)
		return cfg
	}

	// ── HTTP method validation ────────────────────────────────────────────────

	t.Run("HTTP method validation", func(t *testing.T) {
		h := newHandler(t, base())

		t.Run("allows GET", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/home.html")
			if rw.Code == http.StatusMethodNotAllowed {
				t.Error("GET should be allowed")
			}
		})

		t.Run("allows HEAD", func(t *testing.T) {
			rw := doRequest(h, http.MethodHead, "/home.html")
			if rw.Code == http.StatusMethodNotAllowed {
				t.Error("HEAD should be allowed")
			}
		})

		t.Run("rejects POST", func(t *testing.T) {
			rw := doRequest(h, http.MethodPost, "/home.html")
			assertStatus(t, rw, http.StatusMethodNotAllowed)
		})

		t.Run("rejects PUT", func(t *testing.T) {
			rw := doRequest(h, http.MethodPut, "/home.html")
			assertStatus(t, rw, http.StatusMethodNotAllowed)
		})

		t.Run("rejects DELETE", func(t *testing.T) {
			rw := doRequest(h, http.MethodDelete, "/home.html")
			assertStatus(t, rw, http.StatusMethodNotAllowed)
		})

		t.Run("rejects PATCH", func(t *testing.T) {
			rw := doRequest(h, http.MethodPatch, "/home.html")
			assertStatus(t, rw, http.StatusMethodNotAllowed)
		})

		t.Run("rejects OPTIONS", func(t *testing.T) {
			rw := doRequest(h, http.MethodOptions, "/home.html")
			assertStatus(t, rw, http.StatusMethodNotAllowed)
		})

		t.Run("rejects CONNECT", func(t *testing.T) {
			rw := doRequest(h, http.MethodConnect, "/home.html")
			assertStatus(t, rw, http.StatusMethodNotAllowed)
		})

		t.Run("rejects TRACE", func(t *testing.T) {
			rw := doRequest(h, http.MethodTrace, "/home.html")
			assertStatus(t, rw, http.StatusMethodNotAllowed)
		})
	})

	// ── Path traversal protection ─────────────────────────────────────────────

	t.Run("path traversal protection", func(t *testing.T) {
		h := newHandler(t, base())

		t.Run("rejects bare dot-dot", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/..")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("rejects dot-dot in the middle", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir/../home.html")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("rejects percent-encoded dot-dot", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir/%2e%2e/home.html")
			assertStatus(t, rw, http.StatusForbidden)
		})
	})

	// ── Query strings ─────────────────────────────────────────────────────────

	t.Run("query string is ignored", func(t *testing.T) {
		h := newHandler(t, base())

		rw := doRequest(h, http.MethodGet, "/subdir/home.html?foo=bar&baz=1")
		assertStatus(t, rw, http.StatusOK)
		assertBodyEqual(t, rw, "subdir/home.html")
	})

	// ── Content-Type ──────────────────────────────────────────────────────────

	t.Run("Content-Type is set from extension", func(t *testing.T) {
		h := newHandler(t, base())

		t.Run("html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/home.html")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "text/html")
		})

		t.Run("css", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/styles.css")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "text/css")
		})

		t.Run("js", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/app.js")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "text/javascript")
		})

		t.Run("jpg", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/photo.jpg")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "image/jpeg")
		})

		t.Run("png", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/photo.png")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "image/png")
		})

		t.Run("svg", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/photo.svg")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "image/svg+xml")
		})

		t.Run("json", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/data.json")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "application/json")
		})

		t.Run("xml", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/data.xml")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "text/xml")
		})

		t.Run("mp4", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/video.mp4")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "video/mp4")
		})

		t.Run("mkv", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/video.mkv")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "video/x-matroska")
		})

		t.Run("zip", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/archive.zip")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "application/zip")
		})

		t.Run("tar", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/archive.tar")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "application/x-tar")
		})

		// .tar.gz is served by its final extension (.gz).
		t.Run("tar.gz", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/archive.tar.gz")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "application/gzip")
		})

		t.Run("tgz", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/archive.tgz")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "application/x-compressed-tar")
		})

		t.Run("bin", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/raw.bin")
			assertStatus(t, rw, http.StatusOK)
			assertContentType(t, rw, "application/octet-stream")
		})
	})

	// ── Default config ─────────────────────────────────────────

	t.Run("Default config", func(t *testing.T) {
		h := newHandler(t, base())

		t.Run("GET /home.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/home.html")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "home.html")
		})

		t.Run("HEAD /home.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodHead, "/home.html")
			assertStatus(t, rw, http.StatusOK)
			if rw.Body.Len() != 0 {
				t.Errorf("HEAD body should be empty, got %d bytes", rw.Body.Len())
			}
		})

		t.Run("403 when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /subdir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir/")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /subdir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /emptydir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir/")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /emptydir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /nonexistent.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/nonexistent.html")
			assertStatus(t, rw, http.StatusForbidden)
		})
	})

	// ── IndexFile ────────────────────────────────────────────────────────

	t.Run("IndexFile=home.html", func(t *testing.T) {
		cfg := base()
		cfg.IndexFile = "home.html"
		h := newHandler(t, cfg)

		t.Run("serves IndexFile when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "home.html")
		})

		t.Run("serves IndexFile when GET /subdir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir/")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "subdir/home.html")
		})

		t.Run("serves IndexFile when GET /subdir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "subdir/home.html")
		})

		t.Run("403 when GET /emptydir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir/")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /emptydir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir")
			assertStatus(t, rw, http.StatusForbidden)
		})
	})

	// ── FallbackFile ─────────────────────────────────────────────────────

	t.Run("FallbackFile=fallback.html", func(t *testing.T) {
		cfg := base()
		cfg.FallbackFile = "fallback.html"
		h := newHandler(t, cfg)

		t.Run("serves FallbackFile when GET /nonexistent.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/nonexistent.html")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})

		t.Run("serves FallbackFile when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})

		t.Run("serves FallbackFile when GET /subdir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})

		t.Run("serves FallbackFile when GET /emptydir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})
	})

	t.Run("FallbackFile=nonexistent.html", func(t *testing.T) {
		cfg := base()
		cfg.FallbackFile = "nonexistent.html"
		h := newHandler(t, cfg)

		t.Run("403 when GET /nonexistent.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/nonexistent.html")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /subdir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir")
			assertStatus(t, rw, http.StatusForbidden)
		})

		t.Run("403 when GET /emptydir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir")
			assertStatus(t, rw, http.StatusForbidden)
		})
	})

	t.Run("FallbackFile=subdir", func(t *testing.T) {
		cfg := base()
		cfg.FallbackFile = "subdir"
		h := newHandler(t, cfg)

		t.Run("403 when GET /nonexistent.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/nonexistent.html")
			assertStatus(t, rw, http.StatusForbidden)
		})
	})

	// ── ListDir ─────────────────────────────────────────────────────

	t.Run("ListDir=true", func(t *testing.T) {
		cfg := base()
		cfg.ListDir = true
		h := newHandler(t, cfg)

		t.Run("lists when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusOK)
		})

		t.Run("lists when GET /subdir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/subdir/")
			assertStatus(t, rw, http.StatusOK)
		})

		t.Run("lists when GET /emptydir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir/")
			assertStatus(t, rw, http.StatusOK)
		})
	})

	// ── IndexFile + FallbackFile ──────────────────────────────────────────────

	t.Run("IndexFile=home.html and FallbackFile=fallback.html", func(t *testing.T) {
		cfg := base()
		cfg.IndexFile = "home.html"
		cfg.FallbackFile = "fallback.html"
		h := newHandler(t, cfg)

		t.Run("serves IndexFile when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "home.html")
		})

		t.Run("serves FallbackFile when GET /emptydir", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})

		t.Run("serves FallbackFile when GET /nonexistent.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/nonexistent.html")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})
	})

	// ── IndexFile + ListDir ───────────────────────────────────────────────────

	t.Run("IndexFile=home.html and ListDir=true", func(t *testing.T) {
		cfg := base()
		cfg.IndexFile = "home.html"
		cfg.ListDir = true
		h := newHandler(t, cfg)

		t.Run("serves IndexFile when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "home.html")
		})

		t.Run("lists when GET /emptydir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir/")
			assertStatus(t, rw, http.StatusOK)
		})
	})

	// ── FallbackFile + ListDir ────────────────────────────────────────────────

	t.Run("FallbackFile=fallback.html and ListDir=true", func(t *testing.T) {
		cfg := base()
		cfg.FallbackFile = "fallback.html"
		cfg.ListDir = true
		h := newHandler(t, cfg)

		t.Run("serves FallbackFile when GET /nonexistent.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/nonexistent.html")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})

		t.Run("lists when GET /emptydir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir/")
			assertStatus(t, rw, http.StatusOK)
		})
	})

	// ── IndexFile + FallbackFile + ListDir ────────────────────────────────────

	t.Run("IndexFile=home.html, FallbackFile=fallback.html, ListDir=true", func(t *testing.T) {
		cfg := base()
		cfg.IndexFile = "home.html"
		cfg.FallbackFile = "fallback.html"
		cfg.ListDir = true
		h := newHandler(t, cfg)

		t.Run("serves IndexFile when GET /", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "home.html")
		})

		t.Run("lists when GET /emptydir/", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/emptydir/")
			assertStatus(t, rw, http.StatusOK)
		})

		t.Run("serves FallbackFile when GET /nonexistent.html", func(t *testing.T) {
			rw := doRequest(h, http.MethodGet, "/nonexistent.html")
			assertStatus(t, rw, http.StatusOK)
			assertBodyEqual(t, rw, "fallback.html")
		})
	})
}
