# ADR-006: DTOs propios en el adaptador WebSocket

- **Estado:** Aceptada
- **Táctica:** Encapsular — no exponer el modelo interno

## Contexto
Las entidades del dominio no deben depender del formato JSON ni de lo que el cliente necesita ver.

## Decisión
Definir en `adapters/in/websocket/protocol.go` estructuras propias (`roomDTO`, `eventDTO`, ...) con etiquetas JSON y funciones de conversión desde las entidades.

## Consecuencias
- (+) El dominio queda libre de detalles de serialización; el formato del protocolo puede cambiar sin tocar las entidades.
- (−) Duplicación deliberada de campos entre entidades y DTOs (aceptada frente a DRY).

## Alternativas descartadas
- Serializar las entidades directamente: menos código, pero acopla el dominio al contrato externo.
