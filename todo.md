## 0) Базовая инициализация

* [ ] Добавить `Makefile`, `.golangci.yml`, `.editorconfig`, `.gitignore`
* [ ] Структура каталогов (hexagonal):

```
cmd/broadcast-server/main.go
internal/{config,domain,adapters,http,cli,obs}
build/Dockerfile
docker-compose.yml
README.md
```

**Готово, если:** `go run ./cmd/broadcast-server start --port 8080` поднимает пустой HTTP.
**Проверка:** `curl :8080/healthz` → 200.

---

## 1) Конфигурация и флаги

* [ ] ENV/flags: `--port`, `--addr`, `--tls-cert`, `--tls-key`, `--auth-token`, лимиты (`--msg-max-bytes`, `--send-buffer`, `--ping-interval`)
* [ ] CORS origins (список через ENV)
  **Готово, если:** запуск читается из ENV/flags, значения валидируются.
  **Проверка:** `broadcast-server start --port 8090` слушает 8090.

---

## 2) Домейн: Hub

* [ ] Реализовать `Hub` с каналами `register/unregister/broadcast`
* [ ] Карта клиентов, `send`-канал с буфером, backpressure (закрывать медленных)
* [ ] `Run(ctx)` с отменой, счётчики подключений
  **Готово, если:** можно зарегистрировать/отключить клиента и разослать сообщение.
  **Тест:** table-tests на register/broadcast/unregister, `go test -race` зелёный.

---

## 3) WebSocket сервер (adapter http)

* [ ] Маршрут `GET /ws` (chi), апгрейд в WS
* [ ] `SetReadLimit(msgMax)`, `Read/Write` таймауты, `Ping/Pong`
* [ ] JSON-формат: `{"type":"message","sender","text","ts"}`; валидация
* [ ] Ошибки: не-JSON → close (1008), превышение лимита → close
  **Готово, если:** два клиента получают друг друга.
  **Проверка:** `wscat -c ws://localhost:8080/ws?sender=a` и второй — обмен сообщениями.

---

## 4) CLI-клиент (adapter cli)

* [ ] Команда `broadcast-server connect --addr ws://... --name alice`
* [ ] Горутина чтения (WS→stdout), горутина записи (stdin→WS)
* [ ] Закрытие по `SIGINT/SIGTERM` и `EOF`
  **Готово, если:** два CLI-клиента общаются через сервер.
  **Проверка:** открыть два терминала, набрать текст — виден у обоих.

---

## 5) Ограничители и надёжность

* [ ] Rate limit per-connection (token bucket, напр. 20 msg/s)
* [ ] Idle timeout (нет активности N мин — закрыть)
* [ ] Грациозное завершение: stop accept, ждать ≤10s активных клиентов
  **Готово, если:** спамер ограничивается, сервер корректно останавливается.
  **Проверка:** послать >20 сообщений/сек → часть отклоняется/закрытие соединения.

---

## 6) Наблюдаемость

* [ ] Структурные логи (zap): connect/disconnect/reason, sizes, durations
* [ ] Prometheus `/metrics`:

    * `broadcast_connected_clients` (gauge)
    * `broadcast_messages_total{direction=in|out}` (counter)
    * `broadcast_disconnects_total{reason}` (counter)
    * `broadcast_dropped_total` (counter)
* [ ] (dev) pprof `/debug/pprof/`
  **Готово, если:** метрики и профайлы доступны.
  **Проверка:** `curl :8080/metrics` → счётчики видны; `go tool pprof http://:8080/debug/pprof/profile?seconds=10`.

---

## 7) Безопасность (минимум)

* [ ] Опциональный `--auth-token`: требовать `Authorization: Bearer`
* [ ] TLS (`--tls-cert --tls-key`) — включение по флагам
* [ ] CORS allow-list для браузеров
  **Готово, если:** без токена доступ закрыт при включённом режиме; TLS работает.
  **Проверка:** `curl -H "Authorization: Bearer X"` / без заголовка → 401.

---

## 8) Health/Readiness

* [ ] `GET /healthz` — liveness: всегда 200, если процесс жив
* [ ] `GET /readyz` — readiness: hub запущен, при остановке возвращает 503
  **Готово, если:** Kubernetes пробы будут зелёными.
  **Проверка:** `curl :8080/readyz` → 200; при shutdown → 503.

---

## 9) Тестирование

* [ ] Unit: hub (табличные), лимиты, ping/pong
* [ ] Integration: `httptest` + `gorilla/websocket` — connect→send→recv
* [ ] Fuzz: декодер входящих сообщений
* [ ] Bench: broadcast на 100/1000 клиентов (allocs/op, p95 write)
  **Готово, если:** `go test -race -cover ./...` ≥ 80% по домейну/WS, бенчи исполняются.
  **Проверка:** `go test -run . -race -cover ./...`.

---

## 10) Докер и локальная среда

* [ ] Dockerfile (multi-stage, distroless/alpine), неблокирующий ENTRYPOINT
* [ ] `docker-compose.yml` (server + опционально prometheus)
* [ ] Healthcheck для контейнера (`/healthz`)
  **Готово, если:** `docker compose up --build` поднимает сервер и метрики.
  **Проверка:** `docker ps`, `curl localhost:8080/metrics`.

---

## 11) CI/CD (конспект)

* [ ] Jobs: `golangci-lint`, `go test -race -cover`, `govulncheck`
* [ ] Build & push Docker image (tags: `sha`, `latest`)
* [ ] Сбор артефактов: бинарь, SBOM (опц.), coverage отчет
  **Готово, если:** pipeline зелёный на PR/merge.
  **Проверка:** PR в репозиторий → все проверки прошли.

---

## 12) README и UX

* [ ] Установка/запуск: локально и в Docker
* [ ] Примеры: `start`, `connect`, wscat, флаги/ENV
* [ ] Раздел SLO/лимитов по умолчанию, договорённости по код-стайлу/ветвлению
  **Готово, если:** любой разработчик запускает сервис за <5 минут.
  **Проверка:** «чистая» машина/контейнер — повторить шаги README.

---

## 13) Готовность к прод (минимум)

* [ ] Лимиты по памяти/CPU в манифестах; пробы; рестарт-политики
* [ ] Ротация логов/JSON-формат; trace-ids в контексте
* [ ] Настроены дашборды (Grafana) и алерты по метрикам (кол-во клиентов, ошибки)
  **Готово, если:** деплой без ручных шагов, базовые алерты включены.

---

## Make-цели (рекомендуемые)

* [ ] `make lint` / `make test` / `make run-server` / `make run-client`
* [ ] `make docker-build` / `make compose-up` / `make bench`
* [ ] `make vulncheck`

---

## Дополнительно (следующая итерация)

* [ ] Масштабирование по горизонтали: Redis/NATS pubsub между инстансами
* [ ] Комнаты/каналы, история сообщений
* [ ] JWT-аутентификация, ограничение по доменам/реферам для браузеров

---

Готов выдать каркас кода под этот чек-лист (hub + ws + CLI + тесты + Docker).
