package nsr

import (
	"go.uber.org/zap"
)

type Middleware func(next Handler) Handler

type SLogger struct {
	Logger *zap.SugaredLogger
}

func (sl *SLogger) Handler(next Handler) Handler {
	return func(w ResponseWriter, r *Request) error {
		sl.Logger.Infow("handling message on subject %s")
		if err := next(w, r); err != nil {
			sl.Logger.Infow("error: %s", err.Error())
		}
		return nil
	}
}
