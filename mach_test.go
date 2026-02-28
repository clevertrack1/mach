package mach

import (
	"net/http"
	"testing"
)

func TestApp_New(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatal("New() returned nil")
	}
}

func TestApp_Default(t *testing.T) {
	app := Default()
	if len(app.middlewares) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(app.middlewares))
	}
}

func TestApp_RouteMethods(t *testing.T) {
	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			app := New()
			app.Route(method, "/test", func(c *Context) {
				c.Text(200, "ok")
			})

			req := newRequest(method, "/test")
			res := serve(app, req)

			if res.Code != http.StatusOK {
				t.Errorf("%s: expected 200, got %d", method, res.Code)
			}
		})
	}
}

func TestApp_PathParameters(t *testing.T) {
	app := New()
	app.GET("/users/{id}/posts/{postId}", func(c *Context) {
		c.Text(200, "%s-%s", c.Param("id"), c.Param("postId"))
	})

	req := newRequest(http.MethodGet, "/users/123/posts/456")
	res := serve(app, req)

	if res.Body.String() != "123-456" {
		t.Errorf("expected '123-456', got '%s'", res.Body.String())
	}
}

func TestApp_MiddlewareChain(t *testing.T) {
	app := New()
	var order []int

	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 1)
			next.ServeHTTP(w, r)
		})
	})

	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 2)
			next.ServeHTTP(w, r)
		})
	})

	app.GET("/test", func(c *Context) {
		order = append(order, 3)
	})

	req := newRequest(http.MethodGet, "/test")
	res := serve(app, req)

	if res.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", res.Code)
	}

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("wrong order: %v", order)
	}
}

func TestApp_Static(t *testing.T) {
	app := New()
	app.Static("/static", ".")

	req := newRequest(http.MethodGet, "/static/mach.go")
	res := serve(app, req)

	if res.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestApp_404(t *testing.T) {
	app := New()
	app.GET("/exists", func(c *Context) {
		c.Text(200, "found")
	})

	req := newRequest(http.MethodGet, "/notfound")
	res := serve(app, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.Code)
	}
}
