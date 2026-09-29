# ADR-001: Adoptar Arquitectura Hexagonal (Ports & Adapters) para el backend

- **Estado:** Aceptada
- **Táctica:** Modificabilidad — usar un intermediario

## Contexto
El sistema debe poder cambiar de tecnología de persistencia y de protocolo de entrada sin reescribir las reglas de negocio, y el dominio debe poder probarse sin infraestructura. El estilo fue asignado por la cátedra y además encaja con estos requisitos.

## Decisión
Organizar el backend en dominio, puertos y adaptadores: `domain/`, `ports/`, `app/` y `adapters/{in,out}`. Solo `cmd/server/main.go` conoce las implementaciones concretas.

## Consecuencias
- (+) Pruebas del flujo completo con dobles de prueba, sin Redis ni sockets.
- (+) Cambiar Redis o WebSocket implica escribir un adaptador nuevo.
- (−) Más archivos e indirecciones que un CRUD directo; requiere disciplina para no filtrar infraestructura al dominio.

## Alternativas descartadas
- Arquitectura en capas clásica: el dominio terminaría dependiendo del acceso a datos.
- Todo en un `main.go`: rápido de escribir, imposible de probar por partes.
