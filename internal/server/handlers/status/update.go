package status

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
	"github.com/korikhin/pod-sync/internal/models"
	"github.com/korikhin/pod-sync/internal/server"
	"github.com/korikhin/pod-sync/internal/server/middleware/request"
	"github.com/korikhin/pod-sync/internal/storage"
	"github.com/korikhin/pod-sync/internal/watcher"

	"github.com/gorilla/mux"
)

var validator = api.NewValidator()

const queryParamNeedRestart = "need_restart"
const needRestartValue = "true"

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

		p := api.Status{}
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

		status := &models.Status{
			ID: id,
			X:  *p.X,
			Y:  *p.Y,
			Z:  *p.Z,
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

		needRestart := r.URL.Query().Get(queryParamNeedRestart) == needRestartValue
		ops := models.UpdateOperations(status, statusBefore, needRestart)
		wa.QueueOperations(ops)

		httplib.ResponseJSON(w, api.OK("status updated successfully"), http.StatusCreated)
	}

	return http.HandlerFunc(handler)
}
