// Package websocket es el adaptador de entrada (driving adapter): traduce
// mensajes JSON de WebSocket a llamadas del puerto ports.EventService y
// difunde a los clientes los eventos recibidos por ports.EventSubscriber.
package websocket

import (
	"time"

	"contexto-hexagonal/internal/domain"
)

// ---------- Tipos de mensaje (ADR-004, ADR-006) ----------

// Cliente → servidor.
const (
	TypeListRooms  = "list_rooms"
	TypeCreateRoom = "create_room"
	TypeJoin       = "join"
	TypePublish    = "publish"
)

// Servidor → cliente.
const (
	TypeRooms       = "rooms"
	TypeRoomCreated = "room_created"
	TypeJoined      = "joined"
	TypeEvent       = "event"
	TypeError       = "error"
)

// ClientMessage es el sobre de cualquier mensaje entrante. Los campos que
// no aplican a un tipo simplemente quedan vacíos.
type ClientMessage struct {
	Type    string `json:"type"`
	Name    string `json:"name,omitempty"`    // create_room, join
	RoomID  string `json:"roomId,omitempty"`  // join
	Payload string `json:"payload,omitempty"` // publish
}

// ---------- DTOs (nunca se serializan las entidades del dominio) ----------

type RoomDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type ParticipantDTO struct {
	ID     string `json:"id"`
	RoomID string `json:"roomId"`
	Name   string `json:"name"`
}

type EventDTO struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	Author    string    `json:"author"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// ---------- Mensajes servidor → cliente ----------

type RoomsMessage struct {
	Type  string    `json:"type"`
	Rooms []RoomDTO `json:"rooms"`
}

type RoomCreatedMessage struct {
	Type string  `json:"type"`
	Room RoomDTO `json:"room"`
}

type JoinedMessage struct {
	Type        string         `json:"type"`
	Participant ParticipantDTO `json:"participant"`
	History     []EventDTO     `json:"history"`
}

type EventMessage struct {
	Type  string   `json:"type"`
	Event EventDTO `json:"event"`
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ---------- Mapeo dominio → DTO ----------

func toRoomDTO(r domain.Room) RoomDTO {
	return RoomDTO{ID: r.ID, Name: r.Name, CreatedAt: r.CreatedAt}
}

func toParticipantDTO(p domain.Participant) ParticipantDTO {
	return ParticipantDTO{ID: p.ID, RoomID: p.RoomID, Name: p.Name}
}

func toEventDTO(e domain.Event) EventDTO {
	return EventDTO{ID: e.ID, RoomID: e.RoomID, Author: e.Author, Payload: e.Payload, Timestamp: e.Timestamp}
}

func newRoomsMessage(rooms []domain.Room) RoomsMessage {
	dtos := make([]RoomDTO, 0, len(rooms))
	for _, r := range rooms {
		dtos = append(dtos, toRoomDTO(r))
	}
	return RoomsMessage{Type: TypeRooms, Rooms: dtos}
}

func newJoinedMessage(p domain.Participant, history []domain.Event) JoinedMessage {
	dtos := make([]EventDTO, 0, len(history))
	for _, e := range history {
		dtos = append(dtos, toEventDTO(e))
	}
	return JoinedMessage{Type: TypeJoined, Participant: toParticipantDTO(p), History: dtos}
}

func newEventMessage(e domain.Event) EventMessage {
	return EventMessage{Type: TypeEvent, Event: toEventDTO(e)}
}

func newErrorMessage(msg string) ErrorMessage {
	return ErrorMessage{Type: TypeError, Message: msg}
}
