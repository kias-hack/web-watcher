# Архитектура и план реализации web-watcher

## 1. Целевая архитектура

### Слои

| Слой | Назначение |
|------|------------|
| `cmd/app` | Точка входа: флаги, загрузка конфига, сборка зависимостей, запуск/остановка Watchdog |
| `internal/config` | Загрузка и валидация TOML, маппинг в внутренние структуры |
| `internal/domain` | Сущности и бизнес-логика: Service, CheckRule, CheckResult, интерфейсы Checker/Notifier |
| `internal/watchdog` | Оркестратор: шедулинг проверок по интервалам, горутины, вызов Checker и Notifier |
| `internal/infra` | Реализации: HTTP-проверки, Telegram/email, при желании storage/metrics |

Домен не знает про HTTP/Telegram — только про сервисы, правила и результаты. Реализации подключаются через интерфейсы.

---

### Пакеты и сущности

**cmd/app**
- Сборка: config → domain.Service список → HTTP-клиент → Notifier(ы) → Watchdog.
- Запуск `Start()`, обработка сигналов, `Stop(ctx)`.

**internal/config**
- `AppConfig`: global (таймауты, уведомления), `[]ServiceConfig`.
- `ServiceConfig`: Name, URL, Interval, `[]CheckConfig`.
- `CheckConfig`: Type (status_code, body_contains, ssl_not_expired, json_field, …), параметры.
- Ответственность: только загрузка/валидация TOML.

**internal/domain**
- Сущности: `Service`, `CheckRule`, `CheckResult`, `ServiceStatus`, `AlertEvent`.
- Интерфейсы: `ServiceChecker` (CheckService(ctx, Service) ([]CheckResult, error)), `Notifier` (Notify(ctx, AlertEvent) error), опционально `ResultStore`.
- Ответственность: что считается проблемой, когда слать алерт (переход OK→CRIT и т.д.).

**internal/watchdog**
- `Watchdog`: список Service, ServiceChecker, Notifier(ы), опционально ResultStore; ctx, wg.
- Start() — горутина на сервис, тикер по Interval, вызов Checker, сравнение с предыдущим статусом, Notifier при изменении.
- Stop(ctx) — отмена, ожидание wg.

**internal/infra/httpcheck**
- `HTTPServiceChecker`: реализует ServiceChecker, внутри http.Client, маппинг CheckRule → HTTP-проверка (статус, тело, заголовки, SSL).
- Ответственность: один запрос на сервис, выполнение всех правил, возврат []CheckResult.

**internal/infra/notify**
- Реализации `Notifier`: telegram, email; при желании MultiNotifier.
- Ответственность: форматирование AlertEvent в текст, отправка.

**Опционально**
- `internal/infra/storage` — история проверок (файл/SQLite).
- `internal/metrics` или в infra — экспорт в Prometheus.

---

### Типы проверок (чекеры)

- status_code, body_contains, header, max_latency;
- ssl_not_expired, ssl_expires_in_days (WARN/CRIT);
- json_field (путь, ожидаемое значение).
Реализации — в infra, регистрация по типу (map type → handler), без гигантского switch в одном месте.

---

## 2. План задач по этапам

### Этап 0. Подготовка (можно делать сразу)
- [ ] Создать папки пакетов: `internal/domain`, `internal/infra/httpcheck`, `internal/infra/notify`.
- [ ] Ничего не ломать в текущем коде — новые пакеты пока пустые или с заглушками.

**Результат:** дерево пакетов готово к заполнению.

---

### Этап 1. Домен и интерфейсы (делать первым, от него зависит остальное)
- [ ] В `internal/domain`: определить типы `Service`, `CheckRule`, `CheckResult`, `ServiceStatus`, `AlertEvent`.
- [ ] В `internal/domain`: объявить интерфейсы `ServiceChecker` и `Notifier`.
- [ ] (Опционально) Функция/метод в домене: по `[]CheckResult` и предыдущему состоянию решать, нужно ли слать `AlertEvent` (например только при переходе OK→CRIT или CRIT→OK).

**Результат:** ядро модели и контракты; watchdog и инфра потом будут от них зависеть.

**Зависимости:** нет. Можно делать отдельно.

---

### Этап 2. Конфиг под правила
- [ ] Расширить конфиг: у сервиса список правил `[]CheckConfig` с полем `type` и параметрами (expected status, substring, ssl_days и т.д.).
- [ ] Валидация: обязательные поля для каждого типа, допустимые типы.
- [ ] Функция конвертации: `config.ServiceConfig` → `domain.Service` (в config или в отдельном маппере в domain/config).

**Результат:** TOML описывает сервисы и набор проверок; приложение получает `[]domain.Service`.

**Зависимости:** этап 1 (нужны типы CheckRule и т.д.). Можно делать после этапа 1, параллельно с этапом 3 не обязательно — но логичнее после домена.

---

### Этап 3. HTTP-чекер (infra)
- [ ] В `internal/infra/httpcheck`: реализовать `HTTPServiceChecker` (реализует `domain.ServiceChecker`).
- [ ] Внутри: один HTTP-запрос на сервис; для каждой `CheckRule` вызывать соответствующий обработчик (status_code, body_contains, …).
- [ ] Регистрация обработчиков по типу (map или registry), чтобы добавление нового типа не трогало остальной код.
- [ ] Поддержать минимум: status_code, body_contains; затем добавить ssl_not_expired / ssl_expires_in_days, при желании header, max_latency.

**Результат:** по `domain.Service` можно получить `[]CheckResult` через HTTP.

**Зависимости:** этап 1. С этапом 2 можно работать параллельно (для тестов подставлять domain.Service вручную).

---

### Этап 4. Watchdog на домене
- [ ] Переписать `internal/watchdog`: принимает `[]domain.Service`, `ServiceChecker`, `Notifier`(ы), не конфиг напрямую.
- [ ] Start(): для каждого Service — горутина с тикером по Interval; вызов `ServiceChecker.CheckService(ctx, service)`; сохранение/сравнение статуса; при изменении — вызов `Notifier.Notify(AlertEvent)`.
- [ ] Stop(ctx) оставить как сейчас (cancel, wg.Wait с таймаутом).
- [ ] В main собирать зависимости и передавать в Watchdog (config → domain.Service список через маппер из этапа 2, HTTP checker, заглушка Notifier если ещё не готов).

**Результат:** приложение работает через доменные интерфейсы; проверки выполняются, алерты пока заглушкой или логом.

**Зависимости:** этапы 1, 2, 3. Последовательно после них.

---

### Этап 5. Уведомления (Telegram / email)
- [ ] В `internal/infra/notify`: реализовать `Notifier` для Telegram (API бота).
- [ ] В `internal/infra/notify`: реализовать `Notifier` для email (SMTP).
- [ ] Форматирование текста алерта из `AlertEvent` (сервис, правило, статус, детали).
- [ ] Подключить в main (конфиг → создание notifier(ов), передача в Watchdog); при желании MultiNotifier.

**Результат:** при падении/восстановлении админ получает сообщение в TG и/или на почту.

**Зависимости:** этап 1 (интерфейс Notifier), этап 4 (чтобы Watchdog вызывал Notifier). Можно делать параллельно с доработкой чекеров (этап 3), но подключать после этапа 4.

---

### Этап 6. Дополнительные проверки и полировка
- [ ] Добавить проверки: header, max_latency, json_field, ssl_expires_in_days с уровнями WARN/CRIT.
- [ ] systemd: unit-файл, описание в README (запуск, -config путь, логи).
- [ ] По желанию: Prometheus metrics, сохранение истории (storage).

**Результат:** полный набор проверок, удобный запуск под systemd, опционально метрики/история.

**Зависимости:** этапы 1–5. Можно делать по частям и отдельно (например только systemd, или только одна новая проверка).

---

## 3. Порядок и параллелизм

- **Строго последовательно:** 1 → 2 → 4 (домен → конфиг под правила → watchdog на домене). И 3 должен быть до 4 (чекер нужен для watchdog).
- **Параллельно после 1:** этап 2 (конфиг) и этап 3 (HTTP-чекер) можно вести в параллель.
- **После 4:** этап 5 (уведомления) — отдельным блоком, потом этап 6 по желанию.

Кратко: **1 → (2 и 3 параллельно) → 4 → 5 → 6**.

---

## 4. Тесты

**Нужны ли:** да, но минимально и по делу.

- **Имеет смысл:**
  - **domain:** логика “нужен ли алерт при смене статуса” (например переход OK→CRIT да, CRIT→CRIT нет) — чистые функции, легко тестировать.
  - **infra/httpcheck:** тесты с моком HTTP-клиента (httptest.Server или интерфейс с Do()) — проверка, что при таком-то ответе получаем такие CheckResult.
  - **config:** парсинг и валидация (загрузить TOML-файл из testdata, проверить ошибки и заполнение полей).
- **Можно опустить на старте:**
  - Интеграционные тесты (реальный Telegram, SMTP).
  - E2E “поднять приложение и дернуть URL”.
  - Тесты на watchdog с реальным таймером (можно потом добавить с mock time или быстрым интервалом).

Итого: заложить тесты для доменной логики и для HTTP-чекера (с моками) — это даёт пользу и хорошо смотрится в резюме. Остальное можно добавлять по мере времени.
