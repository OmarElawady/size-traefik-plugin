// Package plugindemo a demo plugin.
package plugindemo

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"text/template"

	"github.com/pkg/errors"
)

// Config the plugin configuration.
type Config struct {
	Headers map[string]string `json:"headers,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers: make(map[string]string),
	}
}

// Demo a Demo plugin.
type Demo struct {
	next     http.Handler
	headers  map[string]string
	name     string
	template *template.Template
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}

	return &Demo{
		headers:  config.Headers,
		next:     next,
		name:     name,
		template: template.New("demo").Delims("[[", "]]"),
	}, nil
}

func headerSize(h http.Header) int {
	size := 1
	for k, v := range h {
		size += len(k) + len(v) + 1
	}
	return size
}

func appendToFile(s string) {
	file, err := os.OpenFile("/tmp/temp.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(errors.Wrap(err, "failed to open file"))
	}
	defer file.Close()
	if _, err := file.WriteString(s); err != nil {
		log.Println(errors.Wrap(err, "failed to write to file"))
	}
}

func (a *Demo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wrapper := NewResponseWritrWrapper(rw)

	wr := NewBodyWrapper(req.Body)
	wr.read += uint64(headerSize(req.Header))
	appendToFile(fmt.Sprintf("header size %d: %v\n", wr.read, req.Header))
	req.Body = wr
	a.next.ServeHTTP(wrapper, req)
	appendToFile(fmt.Sprintf("total sent %:\n", wrapper.sent))
	appendToFile(fmt.Sprintf("total received %d\n", wr.read))
}

type BodyWrapper struct {
	body io.ReadCloser
	read uint64
}

func NewBodyWrapper(body io.ReadCloser) *BodyWrapper {
	return &BodyWrapper{
		body: body,
		read: 0,
	}
}

func (b *BodyWrapper) Close() error {
	return b.body.Close()
}

func (b *BodyWrapper) Read(p []byte) (int, error) {
	r, e := b.body.Read(p)
	b.read += uint64(r)
	appendToFile(fmt.Sprintf("read another %d: %s\n", r, string(p)))
	return r, e
}

type ResponseWriterWrapper struct {
	rw   http.ResponseWriter
	sent uint64
}

func NewResponseWritrWrapper(rw http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		rw:   rw,
		sent: 0,
	}
}

func (r *ResponseWriterWrapper) Header() http.Header {
	return r.rw.Header()
}

func (r *ResponseWriterWrapper) Write(d []byte) (int, error) {
	x := uint64(len(d))
	r.sent += x
	appendToFile(fmt.Sprintf("send another %d: %s\n", x, string(d)))
	return r.rw.Write(d)
}

func (r *ResponseWriterWrapper) WriteHeader(statusCode int) {
	r.rw.WriteHeader(statusCode)
}

// Hijack hijacks the connection.
func (r *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.rw.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (r *ResponseWriterWrapper) Flush() {
	if f, ok := r.rw.(http.Flusher); ok {
		f.Flush()
	}
}
