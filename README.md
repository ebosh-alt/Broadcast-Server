# Broadcast-Server
Мини-сервер широковещательной рассылки по WebSocket. Простой, быстрый, готов к Docker.

## Архитектура
* **Hub** (domain): единый цикл `Run` с каналами `register/unregister/broadcast`, карта клиентов, неблокирующая доставка (drop «медленных»).
* **WS-адаптер** (http): `/ws` апгрейд, две горутины на соединение — `reader` (WS→Hub) и `writer` (Hub→WS); ping/pong, таймауты, лимит сообщения (4 KiB).
* **App**: маршруты (`/ws`, `/health`), graceful shutdown.
```
    cmd/broadcast-server/main.go        # точка входа (CLI: start)
    internal/app/app.go                 # wiring, маршруты, graceful shutdown
    internal/domain/hub.go              # Hub: register/unregister/broadcast
    internal/adapters/http/
    ├─ constants.go                     # таймауты/лимиты
    ├─ upgrader.go                      # websocket.Upgrader + CheckOrigin
    ├─ client.go                        # wsClient (send/close)
    ├─ util.go                          # resolveSender, shortRemote
    └─ handler.go                       # ServeWS: /ws, reader/writer
    internal/obs/*                      # (опционально: logger/metrics в будущем)
    build/Dockerfile                    # multi-stage (final: distroless, alpine)
    docker-compose.yml
    Makefile
```

## Команды

```bash
# Локально
make run-server               # старт на :8080
```

```bash
# Здоровье
curl -s localhost:8080/health

# Подключение c именем (zsh экранировать '?')
wscat -c 'ws://localhost:8080/ws?sender=alice'
# Или через заголовок
wscat -H 'X-Sender: bob' -c 'ws://localhost:8080/ws'
```

Пример логов:
```bash
joined: sender="alice" remote=localhost
message: sender="alice" remote=localhost bytes=12
left: sender="alice" remote=localhost
```

## Настройки (по умолчанию)

* Порт: `-port 8080`
* Лимиты: сообщение ≤ 4 KiB, буфер исходящих = 64
* Тайминги: ping 20s / pong 30s / write 10s