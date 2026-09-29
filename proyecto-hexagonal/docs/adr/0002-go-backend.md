# ADR-002: Go como lenguaje y runtime del backend

- **Estado:** Aceptada
- **Táctica:** Rendimiento — concurrencia

## Contexto
El backend mantiene una conexión WebSocket persistente por usuario y debe reenviar eventos a muchas conexiones a la vez.

## Decisión
Usar Go 1.22 con la librería estándar (`net/http`) más `gorilla/websocket` y `go-redis`.

## Consecuencias
- (+) Una goroutine por conexión es barata; el modelo de concurrencia encaja con el problema.
- (+) Compila a un binario estático: imagen Docker mínima.
- (+) Las interfaces se satisfacen implícitamente, lo que simplifica los puertos.
- (−) Manejo de errores más verboso (`if err != nil`) y sin sobrecarga de métodos (ver lecciones aprendidas).

## Alternativas descartadas
- Node.js/Express: válido, pero exige más cuidado con operaciones bloqueantes y no diferenciaba el stack.
- Java/Spring Boot: mayor consumo de memoria y arranque más lento para un servicio pequeño.
