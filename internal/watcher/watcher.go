package watcher

import (
	"log/slog"
	"time"

	"github.com/korikhin/pod-sync/internal/config"
	"github.com/korikhin/pod-sync/internal/lib/logger/sl"
	"github.com/korikhin/pod-sync/internal/models"

	"github.com/korikhin/pod-sync/pkg/deployer"
)

type watcherOptions struct {
	syncInterval time.Duration
}

type Watcher struct {
	log   *slog.Logger
	d     deployer.Deployer
	queue *opQueue
	opts  watcherOptions

	// Канал для отправки команды на завершение
	stopCh chan struct{}

	// Канал для отправки сигнала об успешном завершении
	done chan struct{}
}

func New(log *slog.Logger, d deployer.Deployer, cfg config.Sync) *Watcher {
	log = log.With(sl.Component("sync/watcher"))

	return &Watcher{
		log:    log,
		d:      d,
		queue:  &opQueue{},
		opts:   watcherOptions{syncInterval: cfg.Interval},
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
}

// QueueOperations заносит операции в очередь на выполнение.
func (w *Watcher) QueueOperations(ops []models.PodOperation) {
	if len(ops) == 0 {
		return
	}
	w.queue.add(ops)
	return
}

// Stop отправляет сигнал об остановке и ожидает ответного сигнала об остановке.
//
// TODO: Добавить context.Context остановки.
func (w *Watcher) Stop() {
	close(w.stopCh)
	<-w.done
}

func (w *Watcher) Start() {
	go w.start()
}

// start запускает основной цикл Watcher'а по обработке операций из очереди.
// Обрабатывает операции с подами по таймеру и завершается при вызове Stop.
//
// TODO: Предусмотреть механизм параллельного выполнения операций.
func (w *Watcher) start() {
	defer close(w.done)
	ticker := time.NewTicker(w.opts.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, po := range w.queue.popAll() {
				switch po.Code {
				case models.OpCodeCreate:
					if err := w.d.CreatePod(po.PodID); err != nil {
						w.log.Error("failed to perform operation", sl.PodOperation(po), sl.Error(err))
					} else {
						w.log.Info("operation completed", sl.PodOperation(po))
					}
				case models.OpCodeDelete:
					if err := w.d.DeletePod(po.PodID); err != nil {
						w.log.Error("failed to perform operation", sl.PodOperation(po), sl.Error(err))
					} else {
						w.log.Info("operation completed", sl.PodOperation(po))
					}
				default:
					w.log.Warn("unknown operation", sl.PodOperation(po))
				}
			}
		case <-w.stopCh:
			return
		}
	}
}
