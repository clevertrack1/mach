package mach

import (
	"net/http"
	"testing"
)

func TestGroup_Basic(t *testing.T) {
	app := New()
	group := app.Group("/api")

	group.GET("/users", func(c *Context) {
		c.Text(200, "users")
	})

	req := newRequest(http.MethodGet, "/api/users")
	res := serve(app, req)

	if res.Body.String() != "users" {
		t.Errorf("got %s, want users", res.Body.String())
	}
}

func TestGroup_Nested(t *testing.T) {
	app := New()
	api := app.Group("/api")
	users := api.Group("/users")

	users.GET("/{id}/posts", func(c *Context) {
		c.Text(200, "posts for %s", c.Param("id"))
	})

	req := newRequest(http.MethodGet, "/api/users/123/posts")
	res := serve(app, req)

	if res.Body.String() != "posts for 123" {
		t.Errorf("got %s, want posts for 123", res.Body.String())
	}
}

func TestGroup_Middleware(t *testing.T) {
	app := New()
	order := []int{}

	group := app.Group("/api", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 1)
			next.ServeHTTP(w, r)
		})
	})

	group.GET("/test", func(c *Context) {
		order = append(order, 2)
	})

	req := newRequest(http.MethodGet, "/api/test")
	res := serve(app, req)

	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("wrong order: %v", order)
	}
	if res.Code != http.StatusOK {
		t.Errorf("got %d, want 200", res.Code)
	}
}
