// Package plugindemo a demo plugin.
package plugindemo

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zbus"
	"github.com/threefoldtech/zos/pkg/stubs"
)

// Config the plugin configuration.
type Config struct {
	WID       string `json:"wid,omitempty"`
	MsgBroker string `json:"broker,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		WID: "",
	}
}

// Demo a Demo plugin.
type Demo struct {
	next    http.Handler
	name    string
	WID     string
	gateway *stubs.GatewayStub
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.WID) == 0 {
		return nil, fmt.Errorf("WID (workload id) cannot be empty")
	}
	if len(config.MsgBroker) == 0 {
		config.MsgBroker = "unix:///var/run/redis.sock"
	}
	client, err := zbus.NewRedisClient(config.MsgBroker)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to zbus broker")
	}
	gateway := stubs.NewGatewayStub(client)

	return &Demo{
		WID:     config.WID,
		next:    next,
		name:    name,
		gateway: gateway,
	}, nil
}

func headerSize(h http.Header) int {
	// some headers are not sent from the client
	// like X-Forwarded-Server, should they be counted or not?
	size := 1
	for k, v := range h {
		for _, e := range v {
			size += len(k) + len(e) + 3
		}
	}
	return size
}

func (a *Demo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wrapper := NewResponseWritrWrapper(rw)
	wr := NewBodyWrapper(req.Body)
	wr.read += uint64(headerSize(req.Header))
	log.Debug().Uint64("length", wr.read).Str("content", fmt.Sprintf("%v", req.Header)).Msg("headers")
	wr.read += uint64(len("Host: ") + len(req.Host) + 1) // host is stripped from the headers
	log.Debug().Int("length", len("Host: ")+len(req.Host)+1).Str("content", req.Host).Msg("host header")
	wr.read += uint64(len(req.Proto)) + 1
	log.Debug().Int("length", len(req.Proto)).Str("content", req.Proto).Msg("request protocol")
	wr.read += uint64(len(req.URL.Path)) + 1
	log.Debug().Int("length", len(req.URL.Path)+1).Str("content", req.URL.Path).Msg("request path")
	wr.read += uint64(len(req.Method)) + 1
	log.Debug().Int("length", len(req.Method)).Str("content", req.Method).Msg("request method")

	req.Body = wr
	a.next.ServeHTTP(wrapper, req)

	responseHeaderSize := headerSize(wrapper.Header())
	wrapper.sent += uint64(responseHeaderSize)
	log.Debug().Int("length", len(wrapper.Header())).Str("content", fmt.Sprintf("%v\n", wrapper.Header()))
	wrapper.sent += uint64(len(req.Proto) + 1)
	log.Debug().Int("length", len(req.Proto)).Str("content", req.Proto).Msg("response protocol")
	// it panics for some reason, on the house?
	// wrapper.sent += uint64(len(req.Response.Status) + 1)

	log.Debug().Uint64("value", wr.read).Msg("total received")
	log.Debug().Uint64("value", wrapper.sent).Msg("total sent")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		a.gateway.ReportConsumption(ctx, a.WID, wrapper.sent, wr.read)
	}()
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
	log.Debug().Str("body", string(p)).Int("len", r).Msg("read")
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
	log.Debug().Str("body", string(d)).Uint64("len", x).Msg("sending")
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
