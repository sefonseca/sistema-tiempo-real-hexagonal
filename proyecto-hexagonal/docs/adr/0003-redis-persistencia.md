# ADR-003: Redis como motor de persistencia

- **Estado:** Aceptada
- **Táctica:** Rendimiento — estructuras de datos en memoria adecuadas

## Contexto
Los datos son salas y una secuencia de eventos por sala, con lectura frecuente del historial reciente. La cátedra pide una base de datos clave-valor.

## Decisión
Guardar cada sala como un STRING (JSON), las salas en un ZSET ordenado por fecha y los eventos de cada sala en una LIST con tope de 1000 elementos (`LTRIM`). Activar persistencia AOF en Redis.

## Consecuencias
- (+) Lecturas y escrituras en memoria, con estructuras nativas que evitan modelar tablas.
- (+) La misma tecnología ofrece Pub/Sub (ADR-005).
- (−) Sin integridad referencial ni consultas complejas: el dominio verifica que la sala exista.
- (−) El volumen de datos está limitado por la memoria disponible; por eso el tope de eventos.

## Alternativas descartadas
- PostgreSQL: más adecuado para relaciones complejas, innecesario para este modelo.
- MongoDB: sin ventaja frente a Redis y sin Pub/Sub nativo.
