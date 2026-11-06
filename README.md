# Information Exchange Service

REST API на Go для хранения информации об угрозах по MD5-хешу.

## Запуск

1. Создайте PostgreSQL-базу данных.
2. Скопируйте `.env.example` в `.env` и укажите `DATABASE_URL`.
3. Запустите приложение:

```bash
go run .
```

При запуске выполняется автоматическая миграция таблицы `threats`.

## Архитектура

- `internal/config` — конфигурация и подключение к PostgreSQL;
- `internal/models` — модели GORM;
- `internal/dto` — HTTP-контракты;
- `internal/repository` — операции хранения через интерфейс;
- `internal/service` — валидация и бизнес-логика;
- `internal/handler` — HTTP-обработчики и преобразование ошибок;
- `internal/router` — регистрация маршрутов.

## API

- `POST /info` — создать угрозу;
- `GET /info/{md5}` — получить угрозу;
- `PUT /info/{md5}` — заменить изменяемые поля угрозы.

Полная спецификация находится в `openapi.yaml`.

## Проверка

```bash
go test ./...
```

## Docker и HTTPS/mTLS

Для локального запуска требуются сертификаты в `deploy/certs`. Они не хранятся в Git.

### Генерация сертификатов для локального тестирования

```bash
mkdir -p deploy/certs
cd deploy/certs

# Доверенный CA
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 \
  -out ca.crt -subj "/CN=Information Exchange Local CA"

# Серверный сертификат
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr \
  -subj "/CN=localhost"
printf 'subjectAltName=DNS:localhost\nextendedKeyUsage=serverAuth\n' > server.ext
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 365 -sha256 -extfile server.ext

# Клиентский сертификат
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr \
  -subj "/CN=local-client"
printf 'extendedKeyUsage=clientAuth\n' > client.ext
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 365 -sha256 -extfile client.ext

cd ../..
chmod 600 deploy/certs/*.key
```

Для production не используйте эти локальные сертификаты. Передайте production
сертификаты через `TLS_CERT_DIR`, а приватные ключи храните в защищённом
секрет-хранилище или на хосте с ограниченными правами доступа.

### Запуск

Создайте `.env`:

```dotenv
POSTGRES_PASSWORD=change-me
HTTPS_PORT=8443
```

Запустите локальное окружение:

```bash
docker compose up --build -d
docker compose ps
```

PostgreSQL не публикуется на host. Снаружи доступен только Nginx на
`https://localhost:8443`.

### Проверка TLS и mTLS

Серверный сертификат доверен локальному CA, а клиентский сертификат обязателен:

```bash
curl --cacert deploy/certs/ca.crt \
  --cert deploy/certs/client.crt \
  --key deploy/certs/client.key \
  https://localhost:8443/healthz
```

Запрос без клиентского сертификата должен завершиться ошибкой TLS или HTTP
статусом `495`/`400` в зависимости от версии Nginx:

```bash
curl --cacert deploy/certs/ca.crt https://localhost:8443/healthz
```

Проверка API через mTLS:

```bash
curl --cacert deploy/certs/ca.crt \
  --cert deploy/certs/client.crt \
  --key deploy/certs/client.key \
  https://localhost:8443/info/d41d8cd98f00b204e9800998ecf8427e
```

### Проверка rate limiting

Для GET отправьте больше запросов, чем разрешено лимитом:

```bash
for i in $(seq 1 80); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    --cacert deploy/certs/ca.crt \
    --cert deploy/certs/client.crt \
    --key deploy/certs/client.key \
    https://localhost:8443/healthz
done | sort | uniq -c
```

В результате часть запросов должна получить `429`. Burst допускает короткий
всплеск сверх базовой скорости.

Для production:

```bash
TLS_CERT_DIR=/secure/path/certs \
POSTGRES_PASSWORD='strong-password' \
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

Остановка без удаления данных:

```bash
docker compose down
```
