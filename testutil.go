package mach

import (
	"net/http"
	"net/http/httptest"
)

func newRequest(method, path string) *http.Request {
	return httptest.NewRequest(method, path, nil)
}

func serve(app *App, req *http.Request) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}
