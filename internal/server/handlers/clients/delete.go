package clients

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/korikhin/vortex-assignment/internal/lib/api"
	httplib "github.com/korikhin/vortex-assignment/internal/lib/http"
	"github.com/korikhin/vortex-assignment/internal/lib/logger/sl"
	"github.com/korikhin/vortex-assignment/internal/models"
	"github.com/korikhin/vortex-assignment/internal/server"
	"github.com/korikhin/vortex-assignment/internal/storage"
	"github.com/korikhin/vortex-assignment/internal/watcher"

	"github.com/korikhin/vortex-assignment/internal/server/middleware/request"
)

// Delete удаляет клиента и соответствующий статус.
// Регистрирует операции по удалению активных подов.
func Delete(log *slog.Logger, s server.Storage, wa *watcher.Watcher) http.Handler {
	log = log.With(sl.Component("api/clients"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.clients.Delete"

		log := log.With(
			sl.Operation(op),
			sl.RequestID(request.GetID(r.Context())),
		)

		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			log.Warn("bad request", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrClientNotFound, http.StatusNotFound)
			return
		}

		status, err := s.DeleteClient(context.Background(), id)
		if err != nil {
			if errors.Is(err, storage.ErrClientNotFound) {
				log.Warn("could not delete client", sl.Error(err))
				httplib.ResponseJSON(w, api.ErrClientNotFound, http.StatusNotFound)
				return
			}
			if errors.Is(err, storage.ErrStatusNotFound) {
				log.Warn("could not delete client", sl.Error(err))
				httplib.ResponseJSON(w, api.ErrStatusNotFound, http.StatusNotFound)
				return
			}
			log.Error("failed to delete client", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		ops := models.DeleteOperations(status)
		wa.QueueOperations(ops)

		httplib.ResponseJSON(w, api.Ok(""), http.StatusNoContent)
	}

	return http.HandlerFunc(handler)
}
