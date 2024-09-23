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

### With database

```bash
./cmd/shortener/shortener -d 'postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'
```

### With file storage

```bash
./cmd/shortener/shortener -f './tmp/storage.txt'
```

# API Examples

## Create short URL

### Via text/plain request

```bash
curl -X POST http://localhost:8080 -H "Content-Type: text/plain" -d "https://practicum-yandex.ru"

# Response:
http://localhost:8080/LeKRAJMW
```

### Via application/json request

```bash
curl -i -X POST http://localhost:8080/api/shorten \
    -H "Content-Type: application/json" \
    -d '{"url":"https://practicum-yandex.ru"}'

# Response:
{"result":"http://localhost:8080/TANIJUrQ"}
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
[
    {
        "correlation_id":"mC9g8iasXW",
        "short_url":"http://localhost:8080/TANIJUrQ"
    },
    {
        "correlation_id":"XFADu5Xlkw",
        "short_url":"http://localhost:8080/HdgYTekl"
    }
]
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
