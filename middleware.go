package nsr

import (
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Middleware func(next Handler) Handler

func Logger(l *zap.SugaredLogger) Middleware {
	if l == nil {
		pl, _ := zap.NewProduction()
		l = pl.Sugar()
	}

	return func(next Handler) Handler {
		return func(w ResponseWriter, r *Request) error {
			requestID := ksuid.New()
			l.Info("handling request", "id", requestID, "subject", r.Subject)
			err := next(w, r)

			if err != nil {
				l.Warnw("request finished", "id", requestID, "error", err)
			} else {
				l.Infow("request finished", "id", requestID)
			}
			return err
		}
	}
}
