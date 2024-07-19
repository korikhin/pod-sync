package clients

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/korikhin/vortex-assignment/internal/lib/api"
	httplib "github.com/korikhin/vortex-assignment/internal/lib/http"
	"github.com/korikhin/vortex-assignment/internal/lib/logger/sl"
	"github.com/korikhin/vortex-assignment/internal/server"
	"github.com/korikhin/vortex-assignment/internal/watcher"

	"github.com/korikhin/vortex-assignment/internal/server/middleware/request"
)

// Add создаёт нового клиента и первоначальный статус.
func Add(log *slog.Logger, s server.Storage, wa *watcher.Watcher) http.Handler {
	log = log.With(sl.Component("api/clients"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.clients.Add"

		log := log.With(
			sl.Operation(op),
			sl.RequestID(request.GetID(r.Context())),
		)

		p := api.PayloadClient{}
		if err := httplib.DecodeJSON(r.Body, &p); err != nil {
			log.Error("failed to decode request body", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		if err := api.Validate(p); err != nil {
			log.Warn("bad request", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrBadRequest, http.StatusBadRequest)
			return
		}

		_ /* client */, err := s.AddClient(context.Background(), p)
		if err != nil {
			log.Error("failed to create a client", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		httplib.ResponseJSON(w, api.Ok("client created successfully"), http.StatusCreated)
	}

	return http.HandlerFunc(handler)
}
