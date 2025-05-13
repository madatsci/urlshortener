# URL shortener

[![codecov](https://codecov.io/gh/madatsci/urlshortener/graph/badge.svg?token=CA3XVRAKID)](https://codecov.io/gh/madatsci/urlshortener)

# Development

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Run Database

```bash
docker run --name yandex-practicum-go -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=praktikum -p 5432:5432 -d postgres
```

## Run app

Some examples of how you can run the app (see Configuration below):

### With database

```bash
./cmd/shortener/shortener -d 'postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'
```

### With file storage

```bash
./cmd/shortener/shortener -f './tmp/storage.json'
```

### Configure authentication token

```bash
./cmd/shortener/shortener --token-secret="my_secret_key" --token-duration="1h"
```

## Configuration

App can be configured via flags and/or environment variables. If both flag and environment variable are set for the same parameter, environment variable prevails.

### `-a`, `SERVER_ADDRESS`
Address and port to run server in the form of host:port.

### `-b`, `BASE_URL`
Base URL of the generated short URL.

### `-d`, `DATABASE_DSN`
Database DSN (in case you want to store data in database).

### `-f`, `FILE_STORAGE_PATH`
File storage path (in case you want to store data on disk).

### `--token-secret`, `TOKEN_SECRET_KEY`
Authentication token secret key.

### `--token-duration`, `TOKEN_DURATION`
Authentication token duration (in the format of Golang duration string).

## Migrations

Migrations are implemented with [goose](https://github.com/pressly/goose):

```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable

goose -dir internal/app/store/database/migrations create add_some_column sql
goose -dir internal/app/store/database/migrations up
goose -dir internal/app/store/database/migrations status
```

Migrations are applied automatically when app starts with database DSN provided via flag or environment variable.

# API Examples

## Create short URL

### Via text/plain request

```bash
curl -X POST http://localhost:8080 -H "Content-Type: text/plain" -d "https://practicum-yandex.ru"

# Response:
HTTP/1.1 201 Created
Content-Type: text/plain
Date: Sun, 29 Sep 2024 09:50:29 GMT
Content-Length: 30

http://localhost:8080/fLMxbXUF
```

### Via application/json request

```bash
curl -i -X POST http://localhost:8080/api/shorten \
    -H "Content-Type: application/json" \
    -d '{"url":"https://practicum-yandex.ru"}'

# Response:
HTTP/1.1 201 Created
Content-Type: application/json
Set-Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1cmxzaG9ydGVuZXIiLCJleHAiOjE3Mjc2MDY5NzAsIlVzZXJJRCI6Ijc4MDA3NmU1LTg3M2UtNGQyMC1hM2ZiLWJjNmJlYjVjMGNjNCJ9.LvTerlx2D-jkOuvQqdTKLhrOsS_Op7eglSOLfs3eV4M
Date: Sun, 29 Sep 2024 09:49:35 GMT
Content-Length: 44

{"result":"http://localhost:8080/bnwMHuSR"}
```

### Via application/json batch request

```bash
curl -i -X POST http://localhost:8080/api/shorten/batch \
    -H "Content-Type: application/json" \
    -d '[
        {
            "correlation_id":"mC9g8iasXW",
            "original_url":"https://practicum-yandex.ru"
        },
        {
            "correlation_id":"XFADu5Xlkw",
            "original_url":"http://example.org"
        }
    ]'

# Response:
HTTP/1.1 201 Created
Content-Type: application/json
Set-Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1cmxzaG9ydGVuZXIiLCJleHAiOjE3Mjc2MDY5NzAsIlVzZXJJRCI6ImM1ZDE5ZDZmLTA0YWItNDliOC05NmJlLTVkYjg3YzRhNTgwOSJ9.Na3rNxg9oDxTrQ_h-jsiZbcEd9UCEHrhqrdhWVklW-w
Date: Sun, 29 Sep 2024 09:51:07 GMT
Content-Length: 156

[{"correlation_id":"mC9g8iasXW","short_url":"http://localhost:8080/FgPTdjAI"},{"correlation_id":"XFADu5Xlkw","short_url":"http://localhost:8080/TsHogqxz"}]
```

## Get list of your URLs

### Previously created URLs

```bash
curl -i -X GET -b "auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1cmxzaG9ydGVuZXIiLCJleHAiOjE3Mjc2MDg4MDAsIlVzZXJJRCI6ImI1ZDI4ODdlLTQ0ZWItNGQ4My05OTYzLTI5ZDAxMDBjZTc0ZiJ9.ESKBSqmChOUSpHnxKM42vxANw_atlaArfMtkPWVUndw" http://localhost:8080/api/user/urls

# Response:
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 29 Sep 2024 10:20:26 GMT
Content-Length: 133

[{"short_url":"LduvFKkQ","original_url":"https://practicum-yandex.ru"},{"short_url":"hVKwFYrF","original_url":"http://example.org"}]
```

### No content

```bash
curl -i -X GET -b "auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1cmxzaG9ydGVuZXIiLCJleHAiOjE3Mjc2MDY5NzAsIlVzZXJJRCI6ImM1ZDE5ZDZmLTA0YWItNDliOC05NmJlLTVkYjg3YzRhNTgwOSJ9.Na3rNxg9oDxTrQ_h-jsiZbcEd9UCEHrhqrdhWVklW-w" http://localhost:8080/api/user/urls

# Response:
HTTP/1.1 204 No Content
Date: Sun, 29 Sep 2024 09:53:02 GMT
```

## Use short URL

```bash
curl -i -X GET localhost:8080/LeKRAJMW

# Response:
HTTP/1.1 307 Temporary Redirect
Location: https://practicum-yandex.ru
Date: Mon, 02 Sep 2024 17:52:57 GMT
Content-Length: 0
```

## Delete of your URLs

```bash
curl -i -X DELETE http://localhost:8080/api/user/urls \
    -b "auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1cmxzaG9ydGVuZXIiLCJleHAiOjE3Mjc2MDg4MDAsIlVzZXJJRCI6ImI1ZDI4ODdlLTQ0ZWItNGQ4My05OTYzLTI5ZDAxMDBjZTc0ZiJ9.ESKBSqmChOUSpHnxKM42vxANw_atlaArfMtkPWVUndw" \
    -H "Content-Type: application/json" \
    -d '["LduvFKkQ", "hVKwFYrF"]'

# Response:
HTTP/1.1 202 Accepted
Date: Wed, 02 Oct 2024 13:34:20 GMT
Content-Length: 0
```
