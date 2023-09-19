package nsr

import "github.com/nats-io/nats.go"

type ResponseWriter interface {
	Header(string, string)
	Write([]byte)
}

type Request struct {
	Message nats.Msg
	Header  nats.Header
}

type Handler func(w ResponseWriter, r *Request) error
