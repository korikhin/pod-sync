package watcher

import (
	"sync"

	"github.com/korikhin/vortex-assignment/internal/models"
)

// Данная структура реализует очередь FIFO.
type opQueue struct {
	ops []*models.PodOperation
	mu  sync.Mutex
}

// add добавляет операции в конец очереди.
func (q *opQueue) add(ops []*models.PodOperation) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.ops = append(q.ops, ops...)
}

// popAll очищает очередь и возвращает все элементы.
func (q *opQueue) popAll() []*models.PodOperation {
	q.mu.Lock()
	defer q.mu.Unlock()
	ops := q.ops
	q.ops = make([]*models.PodOperation, 0, cap(ops)/2)
	return ops
}
