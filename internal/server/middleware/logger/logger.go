package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/korikhin/vortex-assignment/internal/lib/logger/sl"

	"github.com/korikhin/vortex-assignment/internal/server/middleware/request"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	log = log.With(sl.Component("middleware/logger"))

	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			log := log.With(
				sl.RequestID(request.GetID(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			log.Info("starting")

			tic := time.Now()
			defer func() {
				tac := time.Since(tic)
				log.Info("completed", sl.Duration(tac))
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handler)
	}
}
