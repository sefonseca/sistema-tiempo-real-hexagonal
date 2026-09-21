// Package ports define los contratos (interfaces) que separan el dominio
// de cualquier tecnología concreta. Nada aquí sabe que existe Go net/http,
// gorilla/websocket ni un cliente de Redis: eso lo saben los adaptadores.
package ports

import "contexto-hexagonal/internal/domain"

// ---------- Puerto de entrada (driving port) ----------
// Lo IMPLEMENTA el dominio (application service) y lo INVOCAN los
// adaptadores de entrada, por ejemplo el handler de WebSocket.
type EventService interface {
	// CreateRoom crea una nueva sala.
	CreateRoom(id, name string) (domain.Room, error)

	// Join agrega un participante a una sala existente.
	Join(participantID, roomID, name string) (domain.Participant, error)

	// PublishEvent valida, persiste y difunde un evento dentro de una sala.
	PublishEvent(id, roomID, author, payload string) (domain.Event, error)

	// History devuelve el histórico de eventos de una sala.
	History(roomID string) ([]domain.Event, error)
}

// ---------- Puertos de salida (driven ports) ----------
// Los IMPLEMENTA un adaptador de salida (ej. repositorio en Redis) y los
// CONSUME el dominio a través de estas interfaces, sin conocer Redis.

// RoomRepository persiste y recupera salas.
type RoomRepository interface {
	Save(r domain.Room) error
	FindByID(id string) (domain.Room, bool, error)
}

// EventRepository persiste y consulta el histórico de eventos.
type EventRepository interface {
	Save(e domain.Event) error
	ListByRoom(roomID string) ([]domain.Event, error)
}

// EventBroadcaster difunde un evento en tiempo real a los participantes
// conectados de una sala (en producción, implementado sobre Redis Pub/Sub).
type EventBroadcaster interface {
	Broadcast(roomID string, e domain.Event) error
}
