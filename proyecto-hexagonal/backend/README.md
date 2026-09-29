# Backend (Go) — Arquitectura Hexagonal

```
cmd/server/main.go            raíz de composición: conecta adaptadores con el dominio
internal/domain/              entidades (Room, Participant, Event) y validaciones
internal/ports/               interfaces: EventService (entrada);
                              RoomRepository, EventRepository, EventBroadcaster, EventSubscriber (salida)
internal/app/                 servicio de aplicación: implementa EventService
internal/adapters/in/websocket/   adaptador de entrada: handler, hub y DTOs
internal/adapters/out/redis/      adaptadores de salida: Store, EventStore, PubSub
internal/platform/id/         generador de identificadores
internal/integration/         prueba end-to-end con Redis real
```

**Regla de dependencia:** `domain`, `ports` y `app` no importan `redis` ni `websocket`. Los adaptadores dependen de los puertos, nunca al revés.

## Variables de entorno

| Variable | Valor por defecto | Descripción |
|---|---|---|
| `PORT` | `8080` | Puerto HTTP/WebSocket |
| `REDIS_ADDR` | `localhost:6379` | Dirección de Redis (en Compose: `redis:6379`) |

## Comandos

```bash
go run ./cmd/server                                   # ejecutar
go test ./...                                         # pruebas unitarias
REDIS_ADDR=localhost:6379 go test -race -count=1 ./...  # incluye integración con Redis
```

## Endpoints

- `GET /ws` — WebSocket (ver protocolo en el README de la raíz)
- `GET /healthz` — responde `ok` si Redis está disponible
