package server

import (
	"context"

	"github.com/korikhin/vortex-assignment/internal/lib/api"
	"github.com/korikhin/vortex-assignment/internal/models"
)

type Storage interface {
	// AddClient создаёт нового клиента и первоначальный статус.
	// Возвращает объект Client и возможную ошибку.
	AddClient(ctx context.Context, p api.PayloadClient) (*models.Client, error)

	// UpdateClient обновляет данные клиента.
	// Возвращает возможную ошибку и статус, если требуется перезагрузка.
	UpdateClient(
		ctx context.Context,
		id int,
		p api.PayloadClient,
		needRestart bool,
	) (*models.Status, error)

	// DeleteClient удаляет клиента.
	// Возвращает соответствующий статус и возможную ошибку.
	DeleteClient(ctx context.Context, id int) (*models.Status, error)

	// UpdateStatus обновляет статус.
	// Возвразает предыдущий статус и возможную ошибку.
	UpdateStatus(ctx context.Context, id int, p api.PayloadStatus) (*models.Status, error)
}
