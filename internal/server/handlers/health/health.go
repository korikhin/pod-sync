package health

import (
	"log/slog"
	"net/http"

	"github.com/korikhin/pod-sync/internal/lib/api"
	httplib "github.com/korikhin/pod-sync/internal/lib/http"
	"github.com/korikhin/pod-sync/internal/lib/logger/sl"
	"github.com/korikhin/pod-sync/internal/server/middleware/request"
)

func New(log *slog.Logger) http.Handler {
	log = log.With(sl.Component("api/health"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			sl.RequestID(request.GetID(r.Context())),
		)

		log.Info("")
		httplib.ResponseJSON(w, api.OK(""), http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
