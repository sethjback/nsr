package nsr

import (
	"bytes"
	"io"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type ResponseWriter interface {
	Header(string, string)
	Write([]byte)
	WriteError(code, description string)
}

type Request struct {
	Subject string
	Date    []byte
	Header  nats.Header
}

type Handler func(w ResponseWriter, r *Request) error

type responseWriter struct {
	headers nats.Header
	body    *bytes.Buffer
	errCode string
	errDesc string
}

func (rw *responseWriter) Header(k, v string) {
	rw.headers.Add(k, v)
}

func (rw *responseWriter) Write(b []byte) {
	rw.body.Write(b)
}

func (rw *responseWriter) WriteError(code, description string) {
	rw.errCode = code
	rw.errDesc = description
}

func handleRequest(mr micro.Request, h Handler) {
	r := &Request{
		Subject: mr.Subject(),
		Date:    mr.Data(),
		Header:  nats.Header(mr.Headers()),
	}

	rw := &responseWriter{
		headers: nats.Header{},
		body:    bytes.NewBuffer(nil),
	}

	err := h(rw, r)
	if err != nil {
		mr.Error("HandlerError", err.Error(), nil)
		return
	}

	if rw.errCode != "" {
		if rw.errDesc == "" {
			rw.errDesc = "unknown"
		}
		mr.Error(rw.errCode, rw.errDesc, nil)
		return
	}

	b, _ := io.ReadAll(rw.body)

	mr.Respond(b, micro.WithHeaders(micro.Headers(rw.headers)))
}
