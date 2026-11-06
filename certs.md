# Сертификаты для локального mTLS

## 1. Подготовка каталога

Из корня проекта создайте каталог сертификатов:

```bash
mkdir -p deploy/certs
cd deploy/certs
```

В результате текущий каталог должен быть:

```text
/Users/maksimblohin/threat-service-blokhin/deploy/certs
```

Файлы из `deploy/certs` подключаются в контейнер Nginx как
`/etc/nginx/certs`. Файлы `*.key`, `*.crt`, `*.csr` и `*.srl` исключены из
Git через `.gitignore`.

## 2. Создание локального доверенного CA

Сначала создайте закрытый ключ CA:

```bash
openssl genrsa -out ca.key 4096
chmod 600 ca.key
```

Создайте самоподписанный сертификат CA:

```bash
openssl req -x509 -new -sha256 -days 365 \
  -key ca.key \
  -out ca.crt \
  -subj "/C=RU/O=Information Exchange/OU=Development/CN=Information Exchange Local CA" \
  -addext "basicConstraints=critical,CA:TRUE,pathlen:1" \
  -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -addext "subjectKeyIdentifier=hash"
```

Проверка сертификата CA:

```bash
openssl x509 -in ca.crt -noout -subject -issuer -serial -dates
openssl x509 -in ca.crt -noout -text | grep -A3 "Basic Constraints"
```

В выводе должно быть:

```text
CA:TRUE
```

`ca.key` является главным секретом этой локальной PKI. Не передавайте его в
Nginx-контейнер и не используйте его как серверный или клиентский ключ.

## 3. Выпуск серверного сертификата

Оставаясь в каталоге `deploy/certs`, создайте конфигурацию расширений
серверного сертификата:

```bash
cat > server.ext <<'EOF'
authorityKeyIdentifier=keyid,issuer
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:localhost,DNS:nginx,IP:127.0.0.1
EOF
```

Создайте закрытый ключ и запрос на подпись:

```bash
openssl genrsa -out server.key 2048
chmod 600 server.key

openssl req -new -sha256 \
  -key server.key \
  -out server.csr \
  -subj "/C=RU/O=Information Exchange/OU=Development/CN=localhost"
```

Подпишите серверный сертификат ранее созданным CA:

```bash
openssl x509 -req -sha256 -days 365 \
  -in server.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out server.crt \
  -extfile server.ext
```

Сертификаты должны соответствовать именам, которые используются при запросах.
Для `https://localhost:8443` нужен `DNS:localhost`. Имя `nginx` добавлено для
обращений внутри Docker-сети.

## 4. Выпуск клиентского сертификата

Создайте конфигурацию расширений клиентского сертификата:

```bash
cat > client.ext <<'EOF'
authorityKeyIdentifier=keyid,issuer
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature
extendedKeyUsage=clientAuth
EOF
```

Создайте закрытый ключ и запрос:

```bash
openssl genrsa -out client.key 2048
chmod 600 client.key

openssl req -new -sha256 \
  -key client.key \
  -out client.csr \
  -subj "/C=RU/O=Information Exchange/OU=Development/CN=local-client"
```

Подпишите клиентский сертификат тем же CA:

```bash
openssl x509 -req -sha256 -days 365 \
  -in client.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out client.crt \
  -extfile client.ext
```

## 5. Проверка подписания сертификатов

Проверьте, что серверный сертификат подписан вашим CA:

```bash
openssl verify -CAfile ca.crt server.crt
```

Ожидаемый результат:

```text
server.crt: OK
```

Проверьте клиентский сертификат:

```bash
openssl verify -CAfile ca.crt client.crt
```

Ожидаемый результат:

```text
client.crt: OK
```

Проверьте, что сертификаты действительно выданы одним CA:

```bash
openssl x509 -in server.crt -noout -issuer -subject
openssl x509 -in client.crt -noout -issuer -subject
```

У обоих сертификатов значение `issuer` должно соответствовать `subject` CA.

## 6. Проверка назначения сертификатов

Проверьте серверный сертификат:

```bash
openssl x509 -in server.crt -noout -text | \
  grep -A2 "Extended Key Usage"
```

В выводе должно быть:

```text
TLS Web Server Authentication
```

Проверьте клиентский сертификат:

```bash
openssl x509 -in client.crt -noout -text | \
  grep -A2 "Extended Key Usage"
```

В выводе должно быть:

```text
TLS Web Client Authentication
```

Проверьте имя `localhost` в серверном сертификате:

```bash
openssl x509 -in server.crt -noout -text | \
  grep -A1 "Subject Alternative Name"
```

В выводе должен присутствовать:

```text
DNS:localhost
```

## 7. Проверка соответствия ключей сертификатам

Сравните публичные ключи сертификатов и соответствующих закрытых ключей:

```bash
openssl x509 -in server.crt -pubkey -noout | \
  openssl pkey -pubin -outform DER | sha256sum
openssl pkey -in server.key -pubout | \
  openssl pkey -pubin -outform DER | sha256sum
```

Два хеша для сервера должны совпадать.

Для клиента:

```bash
openssl x509 -in client.crt -pubkey -noout | \
  openssl pkey -pubin -outform DER | sha256sum
openssl pkey -in client.key -pubout | \
  openssl pkey -pubin -outform DER | sha256sum
```

Два хеша для клиента также должны совпадать.

На macOS вместо `sha256sum`, если эта команда отсутствует, используйте
`shasum -a 256`.

## 8. Какие файлы нужны Nginx и curl

После выпуска сертификатов каталог будет содержать:

```text
deploy/certs/
  ca.crt
  ca.key
  client.crt
  client.csr
  client.ext
  client.key
  server.crt
  server.csr
  server.ext
  server.key
  *.srl
```

Nginx использует только:

```text
ca.crt
server.crt
server.key
```

Для проверки клиента через `curl` используются:

```text
ca.crt
client.crt
client.key
```

Файлы `ca.key`, `*.csr`, `*.ext` и `*.srl` не нужны запущенному Nginx и
должны оставаться только в локальном каталоге администратора PKI.

## 9. Проверка mTLS через Docker Compose

Вернитесь в корень проекта:

```bash
cd ../..
```

Убедитесь, что `.env` заполнен, затем запустите окружение:

```bash
docker compose up --build -d
docker compose ps
```

Запрос с доверенным CA и клиентским сертификатом должен пройти:

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

Проверка без клиентского сертификата должна завершиться ошибкой TLS или
ответом Nginx `400`/`495`:

```bash
curl --cacert deploy/certs/ca.crt \
  https://localhost:8443/healthz
```

Это подтверждает, что `ssl_verify_client on` действительно требует клиентский
сертификат, подписанный доверенным `ca.crt`.

Остановить контейнеры без удаления PostgreSQL volume:

```bash
docker compose down
```
