# ADR-004: WebSocket con mensajes JSON como protocolo de integración

- **Estado:** Aceptada
- **Táctica:** Reducir latencia — canal bidireccional persistente

## Contexto
Los usuarios deben ver los mensajes de los demás al instante, sin recargar ni consultar periódicamente.

## Decisión
Usar WebSocket entre la SPA y el backend, con mensajes JSON tipados (`type`). El protocolo vive en un adaptador de entrada; el dominio no lo conoce.

## Consecuencias
- (+) Latencia mínima y comunicación en ambos sentidos.
- (+) Soporte nativo en navegadores y en Go.
- (−) Hay que gestionar el ciclo de vida de la conexión: se implementaron ping/pong, límite de tamaño y reconexión con backoff en el cliente.
- (−) Algunos proxies requieren configuración para conexiones largas.

## Alternativas descartadas
- REST con polling: mayor latencia y tráfico innecesario.
- Server-Sent Events: solo unidireccional; el cliente igual necesitaría otro canal para enviar.
