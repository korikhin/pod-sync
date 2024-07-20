package handlers

import (
	"log/slog"
	"net/http"

	"github.com/korikhin/vortex-assignment/internal/server"
	"github.com/korikhin/vortex-assignment/internal/server/handlers/clients"
	"github.com/korikhin/vortex-assignment/internal/server/handlers/health"
	"github.com/korikhin/vortex-assignment/internal/server/handlers/status"
	"github.com/korikhin/vortex-assignment/internal/server/middleware/request"
	"github.com/korikhin/vortex-assignment/internal/watcher"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	return mux.NewRouter().PathPrefix("/api").Subrouter()
}

func RegisterHandlers(r *mux.Router, log *slog.Logger, s server.Storage, w *watcher.Watcher) {
	nonEmpty := request.NonEmpty(log)

	// Health
	health := health.New(log)
	r.Handle("/health", health).Methods(http.MethodGet)

	// Clients
	addClient := clients.Add(log, s, w)
	r.Handle("/v1/clients", nonEmpty(addClient)).Methods(http.MethodPost)

	updateClient := clients.Update(log, s, w)
	r.Handle("/v1/clients/{id:[0-9]+}", nonEmpty(updateClient)).Methods(http.MethodPut)

	deleteClient := clients.Delete(log, s, w)
	r.Handle("/v1/clients/{id:[0-9]+}", deleteClient).Methods(http.MethodDelete)

	// Status
	updateStatus := status.Update(log, s, w)
	r.Handle("/v1/status/{id:[0-9]+}", nonEmpty(updateStatus)).Methods(http.MethodPut)

}
