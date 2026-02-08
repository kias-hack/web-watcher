# go-web-watcher

Сервис мониторинга доступности и состояния сайтов (HTTP-проверки, SSL, уведомления).

## Запуск

```bash
go run ./cmd/app -config config/config.toml
```

Флаг `-config` — путь к TOML-конфигу (по умолчанию `config/config.toml`).

## Реализовано

- **Конфиг (TOML):** загрузка файла, `[global]`, `[[services]]`, `prepareService` (имя, interval из global при отсутствии у сервиса).
- **Структуры конфига:** `AppConfig`, `Service`, `GlobalConfig`, `ScrapeConfig`/`AlertConfig`, типы уведомлений (Email, Telegram, Webhook) и SMTP — только структуры, не все поля проброшены в приложение.
- **Watchdog:** старт/стоп, отдельная горутина на каждый сервис, тикер по `interval`, завершение по контексту (SIGTERM/SIGINT), корректное ожидание горутин при Stop.
- **Scraper:** HTTP-запрос по URL, замер латентности, проверка кода ответа, проверка вхождений строк в тело; проверка TLS (наличие сертификата, VerifyHostname, NotBefore/NotAfter, просроченный сертификат, флаг NeedSSLNotify при скором истечении).
- **Service (watchdog):** создание из `config.Service` (URL), структура `ScrapeConfig` (CheckSSL, CheckSSLDate, SSLNotifyPeriod, ExpectedStatusCode, ExpectedBodyContains), хелпер `checkRedirect`.
- **main:** разбор флага `-config`, загрузка конфига, создание и запуск Watchdog, обработка сигналов, shutdown с таймаутом.
- **Тесты:** сценарии Scraper (статус, тело, таймаут, TLS: неверный/верный DNS, просроченный сертификат, скорое истечение, ошибка NotBefore), создание TLS-сервера с кастомным сертификатом.

## Осталось сделать

- **Связка Watchdog ↔ Scraper:** в `scrapeService` сейчас только лог по тикеру; вызывать Scraper, передавать ему `config.Service` (или сконвертированный `watchdog.Service`) и общий `*http.Client`.
- **Маппинг конфига в Scraper:** заполнять `watchdog.Service.ScrapeConfig` из `config.Service`/`GlobalConfig` (expected_status → ExpectedStatusCode, expected_body_contains → ExpectedBodyContains как слайс, verify_ssl, timeout, follow_redirects, max_redirects и т.д.).
- **Уведомления:** реализовать отправку при падении/восстановлении (email по SMTP, Telegram, webhook); хранить состояние «up/down» по каждому сервису и слать алерт только при смене состояния согласно `on_failure`/`on_recovery`.
- **Загрузка SMTP/Telegram из конфига:** добавить секции `[smtp]` и `[telegram]` в `AppConfig`, парсить их из TOML и передавать в код уведомлений.
- **Правки конфига:** исправить опечатки в тегах (`follow_redirects`, `retry_interval`, `retries`), при необходимости поддержать `expected_body_contains` как массив строк в TOML.
- **По желанию:** переменная окружения для пути к конфигу (например `CONFIG_FILE`), флаги/опции уровня логирования.
