package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/korikhin/pod-sync/internal/lib/api"
	httplib "github.com/korikhin/pod-sync/internal/lib/http"
	"github.com/korikhin/pod-sync/internal/lib/logger/sl"
	"github.com/korikhin/pod-sync/internal/server"
	"github.com/korikhin/pod-sync/internal/server/middleware/request"
	"github.com/korikhin/pod-sync/internal/storage"
	"github.com/korikhin/pod-sync/internal/watcher"

	"github.com/gorilla/mux"
)

// Update оновляет данные клиента.
// Регистрирует операции по перезагрузке активных подов, если необходимо.
func Update(log *slog.Logger, s server.Storage, wa *watcher.Watcher) http.Handler {
	log = log.With(sl.Component("api/clients"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.clients.Update"

		log := log.With(
			sl.Operation(op),
			sl.RequestID(request.GetID(r.Context())),
		)

		clientID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			log.Warn("bad request", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrClientNotFound, http.StatusNotFound)
			return
		}

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

		if err := s.UpdateClient(context.Background(), clientID, p); err != nil {
			if errors.Is(err, storage.ErrClientNotFound) {
				log.Warn("could not update client", sl.Error(err))
				httplib.ResponseJSON(w, api.ErrClientNotFound, http.StatusNotFound)
				return
			}
			log.Error("failed to update client", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		httplib.ResponseJSON(w, api.OK("client updated successfully"), http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
