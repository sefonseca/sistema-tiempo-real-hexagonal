# Núcleo hexagonal (dominio + puertos + servicio de aplicación)

Este módulo corresponde a la capa central del hexágono: **no importa nada
de Redis, WebSocket ni HTTP**. Es la parte que, según la Arquitectura
Hexagonal, debe poder probarse y entenderse sin levantar infraestructura.

## Contenido

```
contexto-hexagonal/
├── go.mod
└── internal/
    ├── domain/
    │   ├── entities.go        # Room, Participant, Event + validaciones
    │   └── entities_test.go
    ├── ports/
    │   └── ports.go           # interfaces de entrada y de salida
    └── app/
        ├── event_service.go   # implementa el puerto de entrada (caso de uso)
        └── event_service_test.go
```

## Flujo principal implementado (end-to-end a nivel de dominio)

`CreateRoom` → `Join` → `PublishEvent` (valida → persiste → difunde) → `History`

`PublishEvent` es el caso de uso completo: valida que la sala exista,
valida el contenido del evento con las reglas del dominio, lo guarda a
través de `ports.EventRepository` y lo difunde a través de
`ports.EventBroadcaster`. Si la difusión falla, el evento ya quedó
guardado y el error se reporta sin perder el dato (ver
`TestPublishEvent_EventoQuedaGuardadoAunqueFalleLaDifusion`).

## Cómo se conecta con el resto del proyecto

Los adaptadores que faltan (handler de WebSocket como adaptador de
entrada, y repositorio + Pub/Sub de Redis como adaptadores de salida)
solo necesitan implementar las interfaces de `internal/ports/ports.go`
e inyectarse en `app.NewEventService(...)`. El dominio y el servicio de
aplicación no cambian.

```go
// Ejemplo de cómo se ensamblará en cmd/server/main.go (fuera de esta parte)
svc := app.NewEventService(
    redisRoomRepo,        // implementa ports.RoomRepository
    redisEventRepo,       // implementa ports.EventRepository
    redisPubSubBroadcaster, // implementa ports.EventBroadcaster
    uuid.NewString,
)
```

## Ejecutar

```bash
go build ./...
go test ./... -v
```

Con Go instalado, ambos comandos corren sin ninguna dependencia externa
(no requieren Docker ni Redis) — esa independencia es precisamente lo
que se documentó en la matriz de "Testeabilidad" del análisis arquitectónico.
