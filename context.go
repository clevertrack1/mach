package mach

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ErrEmptyRequestBody = errors.New("request body is empty")
)

// Context adds helpful methods to the ongoing request.
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter

	app *App

	// cached url data
	query        url.Values
	IsFormParsed bool
}

// reset prepares a context to be reused by a new request.
func (ctx *Context) reset(w http.ResponseWriter, r *http.Request) {
	ctx.Response = w
	ctx.Request = r
	ctx.query = nil
	ctx.IsFormParsed = false
}

// Param gets a path parameter by name.
// For example, this returns the value of id from /users/{id}.
func (ctx *Context) Param(name string) string {
	return ctx.Request.PathValue(name)
}

// Query returns a named query parameter.
func (ctx *Context) Query(name string) string {
	// extract all query parameters once
	if ctx.query == nil {
		ctx.query = ctx.Request.URL.Query()
	}

	return ctx.query.Get(name)
}

// DefaultQuery gets query param with default value.
func (ctx *Context) DefaultQuery(name, defaultValue string) string {
	val := ctx.Query(name)
	if val == "" {
		return defaultValue
	}

	return val
}

// Form gets a form value.
func (ctx *Context) Form(name string) string {
	// parse form values only once. its values cached by default once parsed
	if !ctx.IsFormParsed {
		err := ctx.Request.ParseForm()
		if err != nil {
			return ""
		}

		ctx.IsFormParsed = true
	}

	return ctx.Request.FormValue(name)
}

// File gets an uploaded file by key name. The file header containing the file is returned.
func (ctx *Context) File(name string) (*multipart.FileHeader, error) {
	_, header, err := ctx.Request.FormFile(name)

	return header, err
}

// Cookie gets a request cookie by name.
func (ctx *Context) Cookie(name string) (*http.Cookie, error) {
	return ctx.Request.Cookie(name)
}

// GetHeader retrieves a request header by key.
func (ctx *Context) GetHeader(key string) string {
	return ctx.Request.Header.Get(key)
}

// Method returns the request method.
func (ctx *Context) Method() string {
	return ctx.Request.Method
}

// Path retrieves the request path.
func (ctx *Context) Path() string {
	return ctx.Request.URL.Path
}

// ClientIP returns the client IP address.
// Use this if you trust request headers passed to the server (ie: reverse proxy sits before server)
// else use c.Request.RemoteAddr().
func (ctx *Context) ClientIP() string {
	if forwarded := ctx.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")

		return strings.TrimSpace(ips[0])
	}

	if realIP := ctx.Request.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	ip := ctx.Request.RemoteAddr
	if parts := strings.Split(ip, ":"); len(parts) > 1 {
		ip = parts[0]
	}

	return ip
}

// Body reads the request body.
func (ctx *Context) Body() ([]byte, error) {
	return io.ReadAll(ctx.Request.Body)
}

// Context returns the request original context from context.Context.
func (ctx *Context) Context() context.Context {
	return ctx.Request.Context()
}

// binding methods

// DecodeJSON decodes a request body into a struct.
func (ctx *Context) DecodeJSON(data interface{}) error {
	if ctx.Request.Body == nil {
		return ErrEmptyRequestBody
	}

	return json.NewDecoder(ctx.Request.Body).Decode(data)
}

// DecodeXML decodes a request body into a struct.
func (ctx *Context) DecodeXML(data interface{}) error {
	if ctx.Request.Body == nil {
		return ErrEmptyRequestBody
	}

	return xml.NewDecoder(ctx.Request.Body).Decode(data)
}

// response methods

// JSON sends a JSON response.
func (ctx *Context) JSON(status int, data interface{}) error {
	ctx.Response.Header().Set("Content-Type", "application/json")
	ctx.Response.WriteHeader(status)

	return json.NewEncoder(ctx.Response).Encode(data)
}

// XML sends an XML response.
func (ctx *Context) XML(status int, data interface{}) error {
	ctx.SetHeader("Content-Type", "application/xml")
	ctx.Response.WriteHeader(status)

	return xml.NewEncoder(ctx.Response).Encode(data)
}

// Text sends a plain text response.
func (ctx *Context) Text(status int, format string, values ...interface{}) error {
	ctx.SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.Response.WriteHeader(status)

	_, err := fmt.Fprintf(ctx.Response, format, values...)

	return err
}
func (ctx *Context) HTML(status int, html string) error {
	ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
	ctx.Response.WriteHeader(status)

	_, err := ctx.Response.Write([]byte(html))

	return err
}

// Data sends raw bytes.
func (ctx *Context) Data(status int, contentType string, data []byte) error {
	ctx.SetHeader("Content-Type", contentType)
	ctx.Response.WriteHeader(status)

	_, err := ctx.Response.Write(data)

	return err
}

// NoContent sends a response with no body.
func (ctx *Context) NoContent(status int) {
	ctx.Response.WriteHeader(status)
}

// SetHeader sets a response header.
func (ctx *Context) SetHeader(key, value string) {
	ctx.Response.Header().Set(key, value)
}

// GetResponseHeader retrieves a response header by key.
func (ctx *Context) GetResponseHeader(key string) string {
	return ctx.Response.Header().Get(key)
}

// SetCookie sets a response cookie.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(ctx.Response, cookie)
}

// Redirect redirects to a URL.
func (ctx *Context) Redirect(status int, url string) {
	http.Redirect(ctx.Response, ctx.Request, url, status)
}

// utilities

// SaveFile saves an uploaded file to the specified destination path.
func (ctx *Context) SaveFile(file *multipart.FileHeader, path string) error {
	// copy file to destination
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)

	return err
}

// StreamFile streams the content of a file in chunks to the client.
func (ctx *Context) StreamFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	contentType := mime.TypeByExtension(filepath)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	rangeHeader := ctx.GetHeader("Range")

	if rangeHeader == "" {
		ctx.SetHeader("Content-Type", contentType)
		ctx.SetHeader("Content-Length", strconv.FormatInt(stat.Size(), 10))
		ctx.Response.WriteHeader(http.StatusOK)
		_, err = io.Copy(ctx.Response, file)

		return err
	}

	return ctx.serveRange(file, stat, rangeHeader, contentType)
}

// DownloadFile sends a downloadable file response with the specified filename.
func (ctx *Context) DownloadFile(filepath string, downloadName string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if downloadName == "" {
		downloadName = stat.Name()
	}

	contentType := mime.TypeByExtension(filepath)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx.SetHeader("Content-Type", contentType)
	ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%q", downloadName))
	ctx.SetHeader("Content-Length", strconv.FormatInt(stat.Size(), 10))
	ctx.Response.WriteHeader(http.StatusOK)

	_, err = io.Copy(ctx.Response, file)

	return err
}

func (ctx *Context) ServeStatic(dir string) error {
	path := filepath.Join(dir, ctx.Request.URL.Path)

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			path = indexPath
		} else {
			return os.ErrNotExist
		}
	}

	return ctx.StreamFile(path)
}

// serveRange determines the exact range of the file content to serve on the current request context.
func (ctx *Context) serveRange(file *os.File, stat os.FileInfo, rangeHeader, contentType string) error {
	rangePart := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangePart, "-")

	start, end := int64(0), stat.Size()-1
	if len(parts) > 0 && parts[0] != "" {
		start, _ = strconv.ParseInt(parts[0], 10, 64)
	}

	if len(parts) > 1 && parts[1] != "" {
		end, _ = strconv.ParseInt(parts[1], 10, 64)
	}

	if start > end || start >= stat.Size() {
		ctx.SetHeader("Content-Range", fmt.Sprintf("bytes */%d", stat.Size()))
		ctx.Response.WriteHeader(http.StatusRequestedRangeNotSatisfiable)

		return nil
	}

	if end >= stat.Size() {
		end = stat.Size() - 1
	}

	contentLength := end - start + 1

	_, err := file.Seek(start, io.SeekStart)
	if err != nil {
		return err
	}

	ctx.SetHeader("Content-Type", contentType)
	ctx.SetHeader("Content-Length", strconv.FormatInt(contentLength, 10))
	ctx.SetHeader("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, stat.Size()))
	ctx.SetHeader("Accept-Ranges", "bytes")
	ctx.Response.WriteHeader(http.StatusPartialContent)

	_, err = io.CopyN(ctx.Response, file, contentLength)

	return err
}
