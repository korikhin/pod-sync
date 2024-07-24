package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/korikhin/pod-sync/internal/config"
	"github.com/korikhin/pod-sync/internal/lib/api"
	"github.com/korikhin/pod-sync/internal/models"
	"github.com/korikhin/pod-sync/internal/server"
	"github.com/korikhin/pod-sync/internal/storage"

	codes "github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

// New создает и возвращает пул соединений к базе данных PostgreSQL.
// Функция принимает контекст и конфигурацию хранилища.
// Возвращает объект Storage с инициализированным пулом соединений и возможную ошибку.
//
// Процесс работы функции:
//   - Парсит URL подключения из конфигурации.
//   - Настраивает параметры пула соединений.
//   - Создает пул соединений с заданной конфигурацией.
//   - Выполняет пинг базы данных для проверки соединения.
//   - Возвращает объект Storage с инициализированным пулом.
func New(ctx context.Context, cfg config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", op, storage.ErrMalformedConfig, err)
	}

	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MaxConnIdleTime = cfg.IdleTimeout
	poolConfig.MaxConnLifetimeJitter = cfg.LifetimeJitter

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", op, storage.ErrMalformedConfig, err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, sanitizeError(err))
	}

	return &Storage{pool: pool}, nil
}

// Stop закрывает все соединения в пуле и отклоняет новые запросы.
// Блокируется до закрытия всех соединений.
func (s *Storage) Stop() {
	s.pool.Close()
}

func sanitizeError(err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return storage.ErrConnectionTimeout
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Op == "dial" {
			return storage.ErrConnectionDial
		}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if codes.IsConnectionException(pgErr.Code) {
			return storage.ErrConnectionInvalid
		}
		if codes.IsInvalidAuthorizationSpecification(pgErr.Code) {
			return storage.ErrConnectionUnauthorized
		}
	}

	return err
}

var _ server.Storage = (*Storage)(nil)

// AddClient создаёт нового клиента и первоначальный статус.
// Возвращает объект Client и возможную ошибку.
func (s *Storage) AddClient(ctx context.Context, p api.Client) (*models.Client, error) {
	const op = "storage.postgres.AddClient"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	query := `
		insert into watcher.clients (
			name,
			version,
			image,
			cpu,
			mem,
			priority
		) values (
			@name,
			@version,
			@image,
			@cpu,
			@mem,
			@priority
		)
		returning
			id,
			name,
			version,
			image,
			cpu,
			mem,
			priority,
			created_at,
			updated_at;
	`
	args := pgx.NamedArgs{
		"name":     p.Name,
		"version":  p.Version,
		"image":    p.Image,
		"cpu":      p.CPU,
		"mem":      p.Memory,
		"priority": p.Priority,
	}

	client := &models.Client{}
	if err := tx.QueryRow(ctx, query, args).Scan(
		&client.ID,
		&client.Name,
		&client.Version,
		&client.Image,
		&client.CPU,
		&client.Memory,
		&client.Priority,
		&client.CreatedAt,
		&client.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	queryStatus := `
		insert into watcher.status (
			client_id
		) values (
			@client_id
		);
	`
	argsStatus := pgx.NamedArgs{
		"client_id": client.ID,
	}

	if _, err := tx.Exec(ctx, queryStatus, argsStatus); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return client, nil
}

// UpdateClient обновляет данные клиента.
// Возвращает возможную ошибку и статус, если требуется перезагрузка.
func (s *Storage) UpdateClient(ctx context.Context, id int, p api.Client) error {
	const op = "storage.postgres.UpdateClient"

	query := `
		update watcher.clients
		set (
			name,
			version,
			image,
			cpu,
			mem,
			priority,
			updated_at
		) = (
			@name,
			@version,
			@image,
			@cpu,
			@mem,
			@priority,
			timezone('UTC', now())
		)
		where id = @id;
	`
	args := pgx.NamedArgs{
		"id":       id,
		"name":     p.Name,
		"version":  p.Version,
		"image":    p.Image,
		"cpu":      p.CPU,
		"mem":      p.Memory,
		"priority": p.Priority,
	}

	tag, err := s.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrClientNotFound)
	}

	return nil
}

// DeleteClient удаляет клиента.
// Возвращает соответствующий статус и возможную ошибку.
func (s *Storage) DeleteClient(ctx context.Context, id int) (*models.Status, error) {
	const op = "storage.postgres.DeleteClient"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	queryStatus := `
		select
			id,
			"X",
			"Y",
			"Z"
		from watcher.status
		where client_id = @client_id;
	`
	argsStatus := pgx.NamedArgs{
		"client_id": id,
	}

	status := &models.Status{}
	errStatus := tx.QueryRow(ctx, queryStatus, argsStatus).Scan(
		&status.ID,
		&status.X,
		&status.Y,
		&status.Z,
	)

	query := `
		delete from watcher.clients
		where id = @id;
	`
	args := pgx.NamedArgs{
		"id": id,
	}

	if tag, err := tx.Exec(ctx, query, args); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	} else if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrClientNotFound)
	}

	if errStatus != nil {
		if errors.Is(errStatus, pgx.ErrNoRows) {
			status = nil
		} else {
			return nil, fmt.Errorf("%s: %w", op, errStatus)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return status, nil
}

// UpdateStatus обновляет статус.
// Возвразает предыдущий статус и возможную ошибку.
func (s *Storage) UpdateStatus(ctx context.Context, id int, p api.Status) (*models.Status, error) {
	const op = "storage.postgres.UpdateStatus"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	queryGet := `
		select
			"X",
			"Y",
			"Z"
		from watcher.status
		where id = @id;
	`
	argsGet := pgx.NamedArgs{
		"id": id,
	}

	statusBefore := &models.Status{ID: id}
	if err := tx.QueryRow(ctx, queryGet, argsGet).Scan(
		&statusBefore.X,
		&statusBefore.Y,
		&statusBefore.Z,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrStatusNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	queryUpdate := `
		update watcher.status
		set (
			"X",
			"Y",
			"Z"
		) = (
			@X,
			@Y,
			@Z
		)
		where id = @id;
	`
	argsUpdate := pgx.NamedArgs{
		"id": id,
		"X":  p.X,
		"Y":  p.Y,
		"Z":  p.Z,
	}

	tag, err := tx.Exec(ctx, queryUpdate, argsUpdate)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrStatusNotFound)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return statusBefore, nil
}
