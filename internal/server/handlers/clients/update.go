package clients

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/korikhin/vortex-assignment/internal/lib/api"
	httplib "github.com/korikhin/vortex-assignment/internal/lib/http"
	"github.com/korikhin/vortex-assignment/internal/lib/logger/sl"
	"github.com/korikhin/vortex-assignment/internal/models"
	"github.com/korikhin/vortex-assignment/internal/server"
	"github.com/korikhin/vortex-assignment/internal/server/middleware/request"
	"github.com/korikhin/vortex-assignment/internal/storage"
	"github.com/korikhin/vortex-assignment/internal/watcher"

	"github.com/gorilla/mux"
)

const queryParamNeedRestart = "need_restart"

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
			log.Error("failed to decode request body", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		if err := api.Validate(p); err != nil {
			log.Warn("bad request", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrBadRequest, http.StatusBadRequest)
			return
		}

		needRestartStr := r.URL.Query().Get(queryParamNeedRestart)
		needRestart, _ := strconv.ParseBool(needRestartStr)

		status, err := s.UpdateClient(context.Background(), clientID, p, needRestart)
		if err != nil {
			if errors.Is(err, storage.ErrClientNotFound) {
				log.Warn("could not update client", sl.Error(err))
				httplib.ResponseJSON(w, api.ErrClientNotFound, http.StatusNotFound)
				return
			}
			log.Error("failed to update client", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		ops := models.RestartOperations(status)
		wa.QueueOperations(ops)

		httplib.ResponseJSON(w, api.OK("client updated successfully"), http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
