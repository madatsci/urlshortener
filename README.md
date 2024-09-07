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
curl -X POST http://localhost:8080/api/shorten \
    -H "Content-Type: application/json" \
    -d '{"url":"https://practicum-yandex.ru"}'

# Response:
{"result":"http://localhost:8080/TANIJUrQ"}
```

## Use short URL

```bash
curl -I -X GET localhost:8080/LeKRAJMW

# Response:
HTTP/1.1 307 Temporary Redirect
Location: https://practicum-yandex.ru
Date: Mon, 02 Sep 2024 17:52:57 GMT
Content-Length: 0
```
