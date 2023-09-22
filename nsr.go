package nsr

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Router interface {
	Start() error
	Stop() error
	Use(Middleware)
	Path(pattern string, h Handler)
}

type baseRouter struct {
	name       string
	descrption string
	version    string

	nc          *nats.Conn
	srv         micro.Service
	baseSubject string
	middleware  []Middleware
	tree        *node
	eh          func(*micro.NATSError)
	dh          Handler
}

func NewRouter(name string, opts ...Option) (Router, error) {
	r := &baseRouter{name: name, tree: &node{}}

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

	c := micro.Config{
		Name:         b.name,
		Description:  b.descrption,
		Version:      b.version,
		ErrorHandler: b.microError,
	}

	if b.baseSubject == "" {
		b.baseSubject = b.name
	}

	svc, err := micro.AddService(b.nc, c)
	if err != nil {
		return err
	}

	err = svc.AddEndpoint(b.baseSubject, b.microHandler())
	if err != nil {
		return err
	}

	b.srv = svc

	return nil
}

func (b *baseRouter) Stop() error {
	b.srv.Stop()

	if b.nc != nil {
		if err := b.nc.Drain(); err != nil {
			return err
		}
	}

	return nil
}

func (b *baseRouter) Path(pattern string, h Handler) {
	b.tree.addPath(pattern, h)
}

func (b *baseRouter) microError(srv micro.Service, err *micro.NATSError) {
	if b.eh != nil {
		b.eh(err)
	}
}

func (b *baseRouter) Use(m Middleware) {
	b.middleware = append(b.middleware, m)
}

func (b *baseRouter) microHandler() micro.HandlerFunc {
	return micro.HandlerFunc(func(r micro.Request) {
		s := r.Subject()
		h := b.tree.getValue(s)
		if h == nil {
			if b.dh == nil {
				r.Error("NotFound", "subject responder not found", nil)
				return
			}
			h = b.dh
		}

		if len(b.middleware) == 0 {
			handleRequest(r, h)
		} else {
			h := b.middleware[len(b.middleware)-1](h)
			for i := len(b.middleware) - 2; i >= 0; i-- {
				h = b.middleware[i](h)
			}

			handleRequest(r, h)
		}

	})
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

func WithDescription(d string) Option {
	return func(r *baseRouter) error {
		r.descrption = d
		return nil
	}
}

func WithVersion(v string) Option {
	return func(r *baseRouter) error {
		r.version = v
		return nil
	}
}

func WithErrorHandler(f func(*micro.NATSError)) Option {
	return func(r *baseRouter) error {
		r.eh = f
		return nil
	}
}

func WithDefaultHandler(h Handler) Option {
	return func(r *baseRouter) error {
		r.dh = h
		return nil
	}
}
