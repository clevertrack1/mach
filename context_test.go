package mach

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestContext_Query(t *testing.T) {
	app := New()
	app.GET("/test", func(c *Context) {
		err := c.Text(200, "%s", c.Query("key"))
		if err != nil {
			return
		}
	})

	req := newRequest(http.MethodGet, "/test?key=value")
	res := serve(app, req)

	if res.Body.String() != "value" {
		t.Errorf("got %s, want value", res.Body.String())
	}
}

func TestContext_Param(t *testing.T) {
	app := New()
	app.GET("/users/{id}/posts/{postId}", func(c *Context) {
		err := c.Text(200, "%s-%s", c.Param("id"), c.Param("postId"))
		if err != nil {
			return
		}
	})

	req := newRequest(http.MethodGet, "/users/123/posts/456")
	res := serve(app, req)

	if res.Body.String() != "123-456" {
		t.Errorf("got %s, want 123-456", res.Body.String())
	}
}

func TestContext_GetHeader(t *testing.T) {
	app := New()
	app.GET("/test", func(c *Context) {
		err := c.Text(200, "%s", c.GetHeader("X-Custom"))
		if err != nil {
			return
		}
	})

	req := newRequest(http.MethodGet, "/test")
	req.Header.Set("X-Custom", "value")
	res := serve(app, req)

	if res.Body.String() != "value" {
		t.Errorf("got %s, want value", res.Body.String())
	}
}

func TestContext_JSON(t *testing.T) {
	app := New()
	app.GET("/test", func(c *Context) {
		err := c.JSON(200, map[string]string{"message": "hello"})
		if err != nil {
			return
		}
	})

	req := newRequest(http.MethodGet, "/test")
	res := serve(app, req)

	var got map[string]string
	err := json.Unmarshal(res.Body.Bytes(), &got)
	if err != nil {
		return
	}

	if got["message"] != "hello" {
		t.Errorf("got %v, want hello", got)
	}
}

func TestContext_Text(t *testing.T) {
	app := New()
	app.GET("/test", func(c *Context) {
		err := c.Text(200, "Hello %s", "World")
		if err != nil {
			return
		}
	})

	req := newRequest(http.MethodGet, "/test")
	res := serve(app, req)

	if res.Body.String() != "Hello World" {
		t.Errorf("got %s, want Hello World", res.Body.String())
	}
}

func TestContext_NoContent(t *testing.T) {
	app := New()
	app.DELETE("/test", func(c *Context) {
		c.NoContent(http.StatusNoContent)
	})

	req := newRequest(http.MethodDelete, "/test")
	res := serve(app, req)

	if res.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", res.Code)
	}
}

func TestContext_SetCookie(t *testing.T) {
	app := New()
	app.GET("/test", func(c *Context) {
		c.SetCookie(&http.Cookie{Name: "session", Value: "abc"})
		err := c.Text(200, "ok")
		if err != nil {
			return
		}
	})

	req := newRequest(http.MethodGet, "/test")
	res := serve(app, req)

	cookies := res.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" {
		t.Errorf("cookie not set")
	}
}

func TestContext_Redirect(t *testing.T) {
	app := New()
	app.GET("/old", func(c *Context) {
		c.Redirect(http.StatusMovedPermanently, "/new")
	})

	req := newRequest(http.MethodGet, "/old")
	res := serve(app, req)

	if res.Code != http.StatusMovedPermanently || res.Header().Get("Location") != "/new" {
		t.Errorf("redirect not working")
	}
}
