# ADR-005: Difundir eventos mediante Redis Pub/Sub

- **Estado:** Aceptada
- **Táctica:** Escalabilidad — bus de eventos compartido

## Contexto
Si cada instancia del backend difundiera solo a sus propios clientes, dos usuarios conectados a instancias distintas no se verían.

## Decisión
El puerto `EventBroadcaster` publica cada evento en el canal `hexagonal:events` de Redis, y el `Hub` de cada instancia se suscribe mediante el puerto `EventSubscriber` y lo reparte a sus clientes. El evento se guarda **antes** de publicarse.

## Consecuencias
- (+) Se pueden ejecutar varias instancias del backend sin cambiar el dominio.
- (+) Si la difusión falla, el evento ya está persistido.
- (−) Pub/Sub es "como máximo una vez": un cliente desconectado no recibe lo publicado mientras estuvo fuera (lo recupera con el historial al volver a unirse).
- (−) Mejora futura: Redis Streams para reproducción y entrega garantizada.

## Alternativas descartadas
- Difusión solo en memoria: más simple, pero no escala horizontalmente.
- Kafka/RabbitMQ: sobredimensionado para el alcance (YAGNI).
