# Сервис синхронизации подов

Микросервис для синхронизации работы подов. См. [задание](ASSIGNMENT.md).<br>

Целью работы сервиса является синхронизация работы подов с их статусами в базе данных.

## Настройка

В корневой директории создайте [файл](.env.example) `.env`, указав параметры подключения к базе данных:

```ini
POSTGRES_DB=db
POSTGRES_USER=user
POSTGRES_PASSWORD=password
```

<br>

Перечень доступных [параметров](docker-compose.yaml) сервиса со значениями по умолчанию:

- `PSY__SYNC__INTERVAL` — интервал синхронизации (**5m**).
- `PSY__STORAGE__MIN_CONNS` — минимальное количество соединений в пуле (**1**).
- `PSY__STORAGE__MAX_CONNS` — максимальное количество соединений в пуле (**10**).
- `PSY__STORAGE__START_TIMEOUT` — время ожидания при запуске соединения (**30s**).
- `PSY__STORAGE__READ_TIMEOUT` — время ожидания чтения из хранилища (**5s**).
- `PSY__STORAGE__WRITE_TIMEOUT` — время ожидания записи в хранилище (**5s**).
- `PSY__STORAGE__IDLE_TIMEOUT` — время простоя соединения перед закрытием (**30m**).
- `PSY__STORAGE__LIFETIME_JITTER` — случайное отклонение времени жизни соединения (**30s**).
- `PSY__HTTP__READ_TIMEOUT` — время ожидания чтения полного запроса (**5s**).
- `PSY__HTTP__WRITE_TIMEOUT` — время ожидания записи ответа клиенту (**5s**).
- `PSY__HTTP__IDLE_TIMEOUT` — максимальное время простоя соединения (**60s**).
- `PSY__HTTP__SHUTDOWN_TIMEOUT` — время ожидания завершения работы сервера (**10s**).

## Запуск

Сервис запускается с помощью **docker-compose**:

```sh
docker-compose up -d
```

```sh
docker-compose down
```

<br>

В ходе выполнения команды будет развёрнуто два контейнера:

- `watcher` — HTTP сервер (порт `8081`).
- `storage` — база данных PostgreSQL (порт `5433`).

## API

Любой ответ API имеет следующий вид:

```http
000 Status Code
Content-Type: application/json

{
  "status": "status",
  "message": "message"
}
```

Поле `status` принимает значения `ok` или `error`. Поле `message` опционально.

### Здоровье сервиса

```http
GET /api/health
```

```http
200 OK

{
  "status": "ok"
}
```

### Создание клиента

```http
POST /api/v1/clients

{
  "name": "Jimbo",
  "version": 1,
  "image": "...",
  "cpu": "...",
  "mem": "...",
  "priority": 0.26
}
```

```http
201 Created

{
  "status": "ok",
  "message": "client created successfully"
}
```

```http
400 Bad Request
500 Internal Server Error
```

### Обновление клиента

```http
PUT /api/v1/clients/{id:[0-9]+}

{
  "name": "Jimbo",
  "version": 2,
  "image": "...",
  "cpu": "...",
  "mem": "...",
  "priority": 0.26
}
```

```http
200 OK

{
  "status": "ok",
  "message": "client updated successfully"
}
```

```http
400 Bad Request
404 Not Found
500 Internal Server Error
```

### Удаление клиента

```http
DELETE /api/v1/clients/{id:[0-9]+}
```

```http
204 No Content
```

```http
404 Not Found
500 Internal Server Error
```

### Обновление статуса

```http
PUT /api/v1/status/{id:[0-9]+}?need_restart=true

{
  "X": true,
  "Y": false,
  "Z": true
}
```

```http
200 OK

{
  "status": "ok",
  "message": "status updated successfully"
}
```

```http
400 Bad Request
404 Not Found
500 Internal Server Error
```

## Логирование

Логирование осуществляется в `stdout` контейнера `watcher`.<br>

Все операции по созданию/удалению подов сопровождаются соответствующими записями в лог:

```json
{
  "time": "2024-07-19T00:15:22.138231926Z",
  "level": "INFO",
  "msg": "operation completed",
  "component": "sync/watcher",
  "pod_operation": "<CREATE> X-168317"
}
```
