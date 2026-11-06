# Information Exchange Service

REST API на Go для хранения и обмена информацией об угрозах по MD5-хешу
файла.

## Содержание

- [Возможности](#возможности)
- [API](#api)
- [Изменения OpenAPI](#изменения-openapi)
- [Архитектура](#архитектура)
- [Запуск без Docker](#запуск-без-docker)
- [Запуск через Docker Compose](#запуск-через-docker-compose)
- [HTTPS и mTLS](#https-и-mtls)
- [Проверка проекта](#проверка-проекта)
- [CI](#ci)

## Возможности

- создание записи об угрозе;
- поиск угрозы по MD5;
- полная замена изменяемых полей угрозы;
- валидация входных данных;
- PostgreSQL и GORM;
- reverse proxy через Nginx;
- HTTPS с обязательным mTLS;
- rate limiting по IP;
- Docker Compose с отдельными контейнерами API, PostgreSQL и Nginx.

## API

### Общие правила

При прямом запуске API доступен на `http://localhost:8080`.
При запуске через Docker внешний адрес проходит через Nginx:

```text
https://localhost:8443
```

MD5 должен содержать ровно 32 шестнадцатеричных символа. Регистр не имеет
значения: перед сохранением MD5 нормализуется в lowercase.

### `GET /healthz`

Техническая проверка доступности HTTP-сервера.

Ответ `200 OK`:

```json
{
  "status": "ok"
}
```

### `POST /info`

Создаёт новую запись об угрозе.

Пример запроса:

```bash
curl -X POST http://localhost:8080/info \
  -H 'Content-Type: application/json' \
  -d '{
    "md5": "d41d8cd98f00b204e9800998ecf8427e",
    "name": "Trojan.Generic",
    "description": "Generic trojan sample",
    "type": "trojan",
    "severity": "high",
    "status": "active",
    "source": "test-feed",
    "confidence": 95
  }'
```

Успешный ответ: `201 Created`.

Возможные ошибки:

- `400 Bad Request` — некорректный JSON или данные;
- `409 Conflict` — MD5 уже существует;
- `500 Internal Server Error` — внутренняя ошибка.

### `GET /info/{md5}`

Возвращает угрозу по MD5.

```bash
curl http://localhost:8080/info/d41d8cd98f00b204e9800998ecf8427e
```

Успешный ответ: `200 OK`.

Возможные ошибки:

- `400 Bad Request` — некорректный MD5;
- `404 Not Found` — угроза не найдена;
- `500 Internal Server Error` — внутренняя ошибка.

### `PUT /info/{md5}`

Полностью заменяет изменяемые поля существующей угрозы. Поле `name` является
обязательным. Поля, которые не переданы, получают пустое или `unknown`
значение согласно правилам сервиса.

```bash
curl -X PUT http://localhost:8080/info/d41d8cd98f00b204e9800998ecf8427e \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Trojan.Generic.Updated",
    "description": "Updated description",
    "type": "trojan",
    "severity": "critical",
    "status": "active",
    "source": "analyst",
    "confidence": 99
  }'
```

Успешный ответ: `200 OK`.

Возможные ошибки:

- `400 Bad Request` — некорректный MD5 или тело запроса;
- `404 Not Found` — угроза не найдена;
- `500 Internal Server Error` — внутренняя ошибка.

### Формат ошибки

```json
{
  "message": "threat not found",
  "details": "threat not found"
}
```

Полный контракт API находится в [openapi.yaml](./openapi.yaml).

## Изменения OpenAPI

Исходная спецификация описывала только `md5` и `name`. В модель запросов и
ответов добавлены следующие поля:

| Поле | Назначение |
|---|---|
| `description` | Подробное описание угрозы |
| `type` | Тип или семейство угрозы |
| `severity` | Уровень опасности: `low`, `medium`, `high`, `critical`, `unknown` |
| `status` | Состояние записи: `active`, `inactive`, `unknown` |
| `source` | Источник информации |
| `confidence` | Уверенность в классификации от `0` до `100` |
| `first_seen` | Время первого обнаружения |
| `last_seen` | Время последнего обнаружения |
| `created_at` | Время создания записи |
| `updated_at` | Время последнего изменения |

Дополнительно в спецификацию добавлены:

- ограничения длины строк;
- `minLength` и `maxLength` для `name`;
- enum-ограничения для `severity` и `status`;
- диапазон `0..100` для `confidence`;
- формат `date-time` для временных полей;
- комментарии к новым полям непосредственно в YAML;

Технический endpoint `GET /healthz` добавлен в приложение отдельно от
бизнес-эндпоинтов для healthcheck контейнеров.

## Архитектура

```text
HTTP client
    ↓
Nginx: TLS, mTLS, rate limiting, reverse proxy
    ↓
Gin router
    ↓
Handler: HTTP и JSON
    ↓
Service: валидация и бизнес-правила
    ↓
Repository interface
    ↓
GORM
    ↓
PostgreSQL
```

Структура проекта:

```text
.
├── main.go
├── openapi.yaml
├── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
├── internal/
│   ├── config/       # конфигурация и подключение к БД
│   ├── dto/          # HTTP request/response структуры
│   ├── handler/      # HTTP-обработчики
│   ├── models/       # GORM-модели
│   ├── repository/   # хранение данных
│   ├── router/       # регистрация маршрутов
│   └── service/      # бизнес-логика и валидация
└── deploy/
    ├── certs/        # локальные сертификаты, не хранятся в Git
    └── nginx/        # конфигурация Nginx
```

Handlers не обращаются к базе напрямую. Service зависит от интерфейса
Repository, поэтому бизнес-логику можно тестировать без PostgreSQL.

## Запуск без Docker

Требования: Go и доступный PostgreSQL.

Скопируйте пример конфигурации:

```bash
cp .env.example .env
```

Укажите в `.env` корректный `DATABASE_URL`, затем запустите:

```bash
go run .
```

Приложение выполнит автоматическую миграцию таблицы `threats` и будет слушать
адрес из `SERVER_ADDRESS`, по умолчанию `:8080`.

## Запуск через Docker Compose

Локальный `.env` уже содержит параметры разработки. При необходимости измените
пароль:

```dotenv
POSTGRES_DB=threats
POSTGRES_USER=threats_user
POSTGRES_PASSWORD=change-me
POSTGRES_PORT=55432
HTTPS_PORT=8443
```

Запустите сервисы:

```bash
docker compose up --build -d
docker compose ps
```

Будут запущены три контейнера:

- `postgres` — PostgreSQL с volume `postgres_data`;
- `api` — Go-приложение;
- `nginx` — HTTPS reverse proxy.

PostgreSQL в локальном режиме доступен с хоста только на
`127.0.0.1:55432`. Внутри Docker-сети API подключается к
`postgres:5432`. В production внешний порт PostgreSQL отключён.

Остановка без удаления данных:

```bash
docker compose down
```

## HTTPS и mTLS

Для локального тестирования сертификатов используйте пошаговую инструкцию в
[certs.md](./certs.md).

После генерации сертификатов запустите Compose и выполните запрос с клиентским
сертификатом:

```bash
curl --cacert deploy/certs/ca.crt \
  --cert deploy/certs/client.crt \
  --key deploy/certs/client.key \
  https://localhost:8443/healthz
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

Запрос без клиентского сертификата должен быть отклонён:

```bash
curl --cacert deploy/certs/ca.crt \
  https://localhost:8443/healthz
```

Nginx использует TLS 1.2/1.3 и проверяет клиентский сертификат через
`ssl_verify_client on`.

Для production:

```bash
TLS_CERT_DIR=/secure/path/certs \
POSTGRES_PASSWORD='strong-password' \
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

Не добавляйте приватные ключи и сертификаты в Git.

## Rate limiting

Nginx ограничивает запросы на IP-адрес:

| HTTP-метод | Базовый лимит | Burst |
|---|---:|---:|
| `GET` | 60 запросов/минуту | 10 |
| `POST`, `PUT`, `PATCH` | 20 запросов/минуту | 5 |
| `DELETE` | 10 запросов/минуту | 3 |

При превышении лимита возвращается `429 Too Many Requests`.

Проверка GET-лимита:

```bash
for i in $(seq 1 80); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    --cacert deploy/certs/ca.crt \
    --cert deploy/certs/client.crt \
    --key deploy/certs/client.key \
    https://localhost:8443/healthz
done | sort | uniq -c
```

Часть запросов после исчерпания burst должна вернуть `429`.

## Проверка проекта

Из корня проекта:

```bash
gofmt -l .
go vet ./...
go test ./...
git diff --check
docker compose config
```

Если Docker daemon запущен, дополнительно:

```bash
docker compose ps
docker compose logs --tail=100 api postgres nginx
```

## CI

Workflow [ci.yml](./.github/workflows/ci.yml) запускается при каждом `push` и
Pull Request. Он выполняет:

1. проверку форматирования через `gofmt`;
2. статический анализ через `go vet`;
3. тесты через `go test ./...`.

При ошибке любого шага GitHub Actions помечает проверку как failed.
