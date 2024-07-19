package health

import (
	"log/slog"
	"net/http"

	api "github.com/korikhin/vortex-assignment/internal/lib/api"
	httplib "github.com/korikhin/vortex-assignment/internal/lib/http"
	"github.com/korikhin/vortex-assignment/internal/lib/logger/sl"
)

func New(log *slog.Logger) http.Handler {
	log = log.With(sl.Component("api/health"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Info("")
		httplib.ResponseJSON(w, api.Ok(""), http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
