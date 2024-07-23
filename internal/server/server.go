package server

import (
	"context"

	"github.com/korikhin/pod-sync/internal/lib/api"
	"github.com/korikhin/pod-sync/internal/models"
)

type Storage interface {
	// AddClient создаёт нового клиента и первоначальный статус.
	// Возвращает объект Client и возможную ошибку.
	AddClient(ctx context.Context, p api.Client) (*models.Client, error)

	// UpdateClient обновляет данные клиента.
	// Возвращает возможную ошибку и статус, если требуется перезагрузка.
	UpdateClient(ctx context.Context, id int, p api.Client) error

	// DeleteClient удаляет клиента.
	// Возвращает соответствующий статус и возможную ошибку.
	DeleteClient(ctx context.Context, id int) (*models.Status, error)

	// UpdateStatus обновляет статус.
	// Возвразает предыдущий статус и возможную ошибку.
	UpdateStatus(ctx context.Context, id int, p api.Status) (*models.Status, error)
}
