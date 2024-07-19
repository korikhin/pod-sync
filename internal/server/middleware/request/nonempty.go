package request

import (
	"log/slog"
	"net/http"

	"github.com/korikhin/vortex-assignment/internal/lib/api"
	httplib "github.com/korikhin/vortex-assignment/internal/lib/http"
	"github.com/korikhin/vortex-assignment/internal/lib/logger/sl"
)

func NonEmpty(log *slog.Logger) func(next http.Handler) http.Handler {
	log = log.With(sl.Component("middleware/request"))

	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			log := log.With(
				sl.RequestID(GetID(r.Context())),
			)

			if r.Method == http.MethodPost && (r.Body == nil || r.ContentLength == 0) {
				log.Warn("request body is empty")
				httplib.ResponseJSON(w, api.ErrBadRequest, http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handler)
	}
}
