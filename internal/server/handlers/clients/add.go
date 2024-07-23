package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/korikhin/pod-sync/internal/lib/api"
	httplib "github.com/korikhin/pod-sync/internal/lib/http"
	"github.com/korikhin/pod-sync/internal/lib/logger/sl"
	"github.com/korikhin/pod-sync/internal/server"
	"github.com/korikhin/pod-sync/internal/server/middleware/request"
	"github.com/korikhin/pod-sync/internal/watcher"
)

var validator = api.NewValidator()

// Add создаёт нового клиента и первоначальный статус.
func Add(log *slog.Logger, s server.Storage, wa *watcher.Watcher) http.Handler {
	log = log.With(sl.Component("api/clients"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.clients.Add"

		log := log.With(
			sl.Operation(op),
			sl.RequestID(request.GetID(r.Context())),
		)

		p := api.Client{}
		if err := httplib.DecodeJSON(r.Body, &p); err != nil {
			var typeError *json.UnmarshalTypeError
			if errors.As(err, &typeError) {
				log.Warn("bad request", sl.Error(typeError))
				msg := fmt.Sprintf("field %s must be type %s", typeError.Field, typeError.Type)
				httplib.ResponseJSON(w, api.Error(msg), http.StatusBadRequest)
				return
			}
			log.Error("failed to decode request body", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		if err := api.Validate(validator, p); err != nil {
			log.Warn("bad request", sl.Error(err))
			httplib.ResponseJSON(w, api.Error(err.Error()), http.StatusBadRequest)
			return
		}

		_ /* client */, err := s.AddClient(context.Background(), p)
		if err != nil {
			log.Error("failed to create a client", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		httplib.ResponseJSON(w, api.OK("client created successfully"), http.StatusCreated)
	}

	return http.HandlerFunc(handler)
}
