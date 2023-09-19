package nsr

import (
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Router interface {
	Start() error
	Stop() error
}

type baseRouter struct {
	nc          *nats.Conn
	sub         *nats.Subscription
	baseSubject string
}

func NewRouter(opts ...Option) (Router, error) {
	r := &baseRouter{}

	for _, o := range opts {
		err := o(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (b *baseRouter) Start() error {
	if b.nc == nil {
		return errors.New("no nats connection")
	}

	sub, err := b.nc.Subscribe(b.baseSubject, func(msg *nats.Msg) {
		fmt.Println("message received", msg.Subject)
	})

	if err != nil {
		return err
	}

	b.sub = sub

	return nil
}

func (b *baseRouter) Stop() error {
	if b.nc != nil {
		if err := b.nc.Drain(); err != nil {
			return err
		}
	}

	return nil
}

type Option func(r *baseRouter) error

func WithNatsConnection(nc *nats.Conn) Option {
	return func(r *baseRouter) error {
		r.nc = nc
		return nil
	}
}

func WithBaseSubject(sub string) Option {
	return func(r *baseRouter) error {
		r.baseSubject = sub
		return nil
	}
}
