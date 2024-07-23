package clients

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/korikhin/pod-sync/internal/lib/api"
	httplib "github.com/korikhin/pod-sync/internal/lib/http"
	"github.com/korikhin/pod-sync/internal/lib/logger/sl"
	"github.com/korikhin/pod-sync/internal/models"
	"github.com/korikhin/pod-sync/internal/server"
	"github.com/korikhin/pod-sync/internal/server/middleware/request"
	"github.com/korikhin/pod-sync/internal/storage"
	"github.com/korikhin/pod-sync/internal/watcher"

	"github.com/gorilla/mux"
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
			log.Error("failed to delete client", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}
		if status == nil {
			log.Warn("client deleted, but status not found")
		}

		ops := models.DeleteOperations(status)
		wa.QueueOperations(ops)

		httplib.ResponseJSON(w, api.OK(""), http.StatusNoContent)
	}

	return http.HandlerFunc(handler)
}
