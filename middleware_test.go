package mach

import (
	"net/http"
	"testing"
)

func TestLogger(t *testing.T) {
	app := New()
	app.Use(Logger())
	app.GET("/test", func(c *Context) {
		c.Text(200, "ok")
	})

	req := newRequest(http.MethodGet, "/test")
	res := serve(app, req)

	if res.Code != http.StatusOK {
		t.Errorf("got %d, want 200", res.Code)
	}
}

func TestRecovery(t *testing.T) {
	app := New()
	app.Use(Recovery())
	app.GET("/panic", func(c *Context) {
		panic("test panic")
	})

	req := newRequest(http.MethodGet, "/panic")
	res := serve(app, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", res.Code)
	}
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name        string
		config      CORSConfig
		origin      string
		checkHeader string
		checkValue  string
	}{
		{
			name: "allow all",
			config: CORSConfig{
				AllowOrigins: []string{"*"},
			},
			origin:      "http://example.com",
			checkHeader: "Access-Control-Allow-Origin",
			checkValue:  "*",
		},
		{
			name: "specific origin allowed",
			config: CORSConfig{
				AllowOrigins: []string{"http://example.com"},
			},
			origin:      "http://example.com",
			checkHeader: "Access-Control-Allow-Origin",
			checkValue:  "http://example.com",
		},
		{
			name: "specific origin not allowed",
			config: CORSConfig{
				AllowOrigins: []string{"http://example.com"},
			},
			origin:      "http://evil.com",
			checkHeader: "Access-Control-Allow-Origin",
			checkValue:  "",
		},
		{
			name: "with credentials",
			config: CORSConfig{
				AllowOrigins:     []string{"http://example.com"},
				AllowCredentials: true,
			},
			origin:      "http://example.com",
			checkHeader: "Access-Control-Allow-Credentials",
			checkValue:  "true",
		},
		{
			name: "with max age",
			config: CORSConfig{
				AllowOrigins: []string{"http://example.com"},
				MaxAge:       3600,
			},
			origin:      "http://example.com",
			checkHeader: "Access-Control-Max-Age",
			checkValue:  "3600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()
			app.Use(CORSWithConfig(tt.config))
			app.GET("/test", func(c *Context) {
				c.Text(200, "ok")
			})

			req := newRequest(http.MethodGet, "/test")
			req.Header.Set("Origin", tt.origin)
			res := serve(app, req)

			if got := res.Header().Get(tt.checkHeader); got != tt.checkValue {
				t.Errorf("got %s, want %s", got, tt.checkValue)
			}
		})
	}
}

func TestCORS_Preflight(t *testing.T) {
	app := New()
	app.Use(CORS([]string{"http://example.com"}))

	req := newRequest(http.MethodOptions, "/test")
	req.Header.Set("Origin", "http://example.com")
	res := serve(app, req)

	if res.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", res.Code)
	}
}
