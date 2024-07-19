package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/korikhin/vortex-assignment/internal/config"
	"github.com/korikhin/vortex-assignment/internal/lib/logger/sl"
	"github.com/korikhin/vortex-assignment/internal/server/handlers"

	storage "github.com/korikhin/vortex-assignment/internal/storage/postgres"
	"github.com/korikhin/vortex-assignment/internal/watcher"
	deployer "github.com/korikhin/vortex-assignment/pkg/deployer/mocks"

	"github.com/korikhin/vortex-assignment/internal/server/middleware/logger"
	"github.com/korikhin/vortex-assignment/internal/server/middleware/request"
)

func main() {
	cfg := config.MustLoad()

	log := sl.New()
	log.Debug("debug messages are enabled")

	// Конфигурация хранилища
	storage, err := storage.New(context.Background(), cfg.Storage)
	if err != nil {
		log.Error("failed to initialize storage", sl.Error(err))
		os.Exit(1)
	}

	deployer := deployer.New()

	// Сервис реализующий синхронизацию статусов
	watcher := watcher.NewWatcher(log, deployer, cfg.Sync)
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
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("failed to start the server", sl.Error(err))
				os.Exit(1)
			}
		}
	}()

	// Ожидание сигнала прерывания работы сервиса
	shutdownSignal := <-shutdown
	log.Info("recieved shutdown signal", sl.Signal(shutdownSignal))
	log.Info("stopping service...")

	watcher.Stop() // Ожидаем остановку
	storage.Stop() // Ожидаем закрытия всех соединений

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("error occured while stopping the server", sl.Error(err))
		os.Exit(1)
	}

	log.Info("service stopped successfully")
	os.Exit(0)
}
