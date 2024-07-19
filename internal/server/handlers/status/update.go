package status

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
	"github.com/korikhin/vortex-assignment/internal/storage"
	"github.com/korikhin/vortex-assignment/internal/watcher"

	"github.com/korikhin/vortex-assignment/internal/server/middleware/request"

	"github.com/gorilla/mux"
)

// Update обновляет статус.
// Регистрирует соответствующие операции с подами.
func Update(log *slog.Logger, s server.Storage, wa *watcher.Watcher) http.Handler {
	log = log.With(sl.Component("api/status"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.status.Update"

		log := log.With(
			sl.Operation(op),
			sl.RequestID(request.GetID(r.Context())),
		)

		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			log.Warn("bad request", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrStatusNotFound, http.StatusNotFound)
			return
		}

		p := api.PayloadStatus{}
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

		status := &models.Status{
			ID:   id,
			VWAP: *p.VWAP,
			TWAP: *p.TWAP,
			HFT:  *p.HFT,
		}

		statusBefore, err := s.UpdateStatus(context.Background(), id, p)
		if err != nil {
			if errors.Is(err, storage.ErrStatusNotFound) {
				log.Warn("could not update status", sl.Error(err))
				httplib.ResponseJSON(w, api.ErrStatusNotFound, http.StatusNotFound)
				return
			}
			log.Error("failed to update status", sl.Error(err))
			httplib.ResponseJSON(w, api.ErrInternal, http.StatusInternalServerError)
			return
		}

		ops := models.UpdateOperations(status, statusBefore)
		wa.QueueOperations(ops)

		httplib.ResponseJSON(w, api.Ok("status updated successfully"), http.StatusCreated)
	}

	return http.HandlerFunc(handler)
}
