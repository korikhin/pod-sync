package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/korikhin/pod-sync/internal/config"
	"github.com/korikhin/pod-sync/internal/lib/logger/sl"
	"github.com/korikhin/pod-sync/internal/server/handlers"
	"github.com/korikhin/pod-sync/internal/server/middleware/logger"
	"github.com/korikhin/pod-sync/internal/server/middleware/request"
	storage "github.com/korikhin/pod-sync/internal/storage/postgres"
	"github.com/korikhin/pod-sync/internal/watcher"

	deployer "github.com/korikhin/pod-sync/pkg/deployer/mocks"
)

func main() {
	cfg := config.MustLoad()

	log := sl.New()
	log.Debug("debug messages are enabled")

	// Конфигурация хранилища
	storage, err := storage.New(context.Background(), cfg.Storage)
	if err != nil {
		log.Error("failed to initialize the storage", sl.Error(err))
		os.Exit(1)
	}

	deployer := deployer.New()

	// Сервис реализующий синхронизацию статусов
	watcher := watcher.New(log, deployer, cfg.Sync)
	watcher.Start()

	// Конфигурация HTTP сервера
	router := handlers.NewRouter()
	router.Use(
		request.ID(),
		logger.New(log),
	)
	handlers.RegisterHandlers(router, log, storage, watcher)
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		IdleTimeout:    cfg.HTTP.IdleTimeout,
		MaxHeaderBytes: 4 << 10, // 4 KiB
	}

	// Запуск сервера
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("failed to start the server", sl.Error(err))
				cancel()
			}
		}
	}()

	// Ожидание сигнала прерывания работы сервиса
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-shutdown:
		log.Info("recieved shutdown signal", sl.Signal(s))
	case <-ctx.Done():
	}

	log.Info("stopping service...")

	shCtx, shCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shCancel()

	if err := server.Shutdown(shCtx); err != nil {
		log.Error("error occurred while stopping the server", sl.Error(err))
	}

	watcher.Stop() // Ожидаем остановку
	storage.Stop() // Ожидаем закрытия всех соединений

	log.Info("service stopped successfully")
}
