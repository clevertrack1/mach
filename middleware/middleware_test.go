package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clevertrack1/mach"
)

func TestRequestID(t *testing.T) {
	app := mach.New()
	app.Use(RequestID())
	app.GET("/test", func(c *mach.Context) {
		err := c.Text(200, "ok")
		if err != nil {
			return
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestRequestID_UseExisting(t *testing.T) {
	app := mach.New()
	app.Use(RequestID())
	app.GET("/test", func(c *mach.Context) {
		err := c.Text(200, "ok")
		if err != nil {
			return
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "existing-id" {
		t.Error("should use existing request ID")
	}
}

func TestGzip(t *testing.T) {
	app := mach.New()
	app.Use(Gzip())
	app.GET("/test", func(c *mach.Context) {
		err := c.Text(200, "hello world")
		if err != nil {
			return
		}
	})

	t.Run("with gzip accept", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("gzip encoding should be set")
		}
	})

	t.Run("without gzip accept", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Header().Get("Content-Encoding") == "gzip" {
			t.Error("gzip should not be set")
		}
	})
}
