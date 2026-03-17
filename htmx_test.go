package mach

import (
	"net/http"
	"strings"
	"testing"
)

func TestHTMX_RequestHeaders(t *testing.T) {
	app := New()
	app.GET("/htmx", func(c *Context) {
		h := c.HTMX()
		if !h.IsHTMX() {
			t.Errorf("expected IsHTMX to be true")
		}
		if h.Target() != "main-content" {
			t.Errorf("expected Target to be main-content, got %s", h.Target())
		}
		if h.Trigger() != "search-button" {
			t.Errorf("expected Trigger to be search-button, got %s", h.Trigger())
		}
		if h.CurrentURL() != "http://localhost/page" {
			t.Errorf("expected CurrentURL to be http://localhost/page, got %s", h.CurrentURL())
		}
		c.Text(200, "ok")
	})

	req := newRequest(http.MethodGet, "/htmx")
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "main-content")
	req.Header.Set("HX-Trigger", "search-button")
	req.Header.Set("HX-Current-URL", "http://localhost/page")

	serve(app, req)
}

func TestHTMX_ResponseHeaders(t *testing.T) {
	app := New()
	app.GET("/htmx-res", func(c *Context) {
		h := c.HTMX()
		h.PushURL("/new-url")
		h.Retarget("#new-target")
		h.TriggerResponse("event1")
		h.TriggerResponse(map[string]string{"event2": "val2"})
		c.NoContent(200)
	})

	req := newRequest(http.MethodGet, "/htmx-res")
	res := serve(app, req)

	if res.Header().Get("HX-Push-Url") != "/new-url" {
		t.Errorf("expected HX-Push-Url to be /new-url, got %s", res.Header().Get("HX-Push-Url"))
	}
	if res.Header().Get("HX-Retarget") != "#new-target" {
		t.Errorf("expected HX-Retarget to be #new-target, got %s", res.Header().Get("HX-Retarget"))
	}
	// Note: consecutive calls to TriggerResponse should now merge or append.
	// Since both "event1" and {"event2": "val2"} are mixed types, it should result in "event1, {"event2": "val2"}"
	expected := `event1, {"event2":"val2"}`
	if res.Header().Get("HX-Trigger") != expected {
		t.Errorf("expected HX-Trigger to be %s, got %s", expected, res.Header().Get("HX-Trigger"))
	}
}

func TestHTMX_TriggerMerge(t *testing.T) {
	app := New()
	app.GET("/merge", func(c *Context) {
		h := c.HTMX()
		h.TriggerResponse(map[string]string{"event1": "val1"})
		h.TriggerResponse(map[string]string{"event2": "val2"})
		c.NoContent(200)
	})

	req := newRequest(http.MethodGet, "/merge")
	res := serve(app, req)

	// Since they are both JSON objects, they should be merged.
	// Map iteration is random, so we check for both keys in the JSON.
	got := res.Header().Get("HX-Trigger")
	if !strings.Contains(got, `"event1":"val1"`) || !strings.Contains(got, `"event2":"val2"`) {
		t.Errorf("expected merged JSON, got %s", got)
	}
}

func TestHTMX_Refresh(t *testing.T) {
	app := New()
	app.GET("/refresh", func(c *Context) {
		c.HTMX().Refresh()
		c.NoContent(200)
	})

	req := newRequest(http.MethodGet, "/refresh")
	res := serve(app, req)

	if res.Header().Get("HX-Refresh") != "true" {
		t.Errorf("expected HX-Refresh to be true, got %s", res.Header().Get("HX-Refresh"))
	}
}
