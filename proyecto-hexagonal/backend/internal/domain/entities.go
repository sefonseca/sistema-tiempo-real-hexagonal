// Package domain contiene el núcleo del negocio: entidades y reglas.
// No importa nada de frameworks, drivers de base de datos ni de red:
// eso es justamente lo que exige Arquitectura Hexagonal (Ports & Adapters).
package domain

import (
	"errors"
	"strings"
	"time"
)

// Errores de negocio propios del dominio (no HTTP, no WebSocket, no Redis).
var (
	ErrEmptyRoomName    = errors.New("domain: el nombre de la sala no puede estar vacío")
	ErrEmptyParticipant = errors.New("domain: el nombre del participante no puede estar vacío")
	ErrEmptyPayload     = errors.New("domain: el evento no puede tener contenido vacío")
	ErrPayloadTooLong   = errors.New("domain: el evento supera el tamaño máximo permitido")
	ErrRoomNotFound     = errors.New("domain: la sala no existe")
)

const maxPayloadLength = 2000

// Room (Entidad de negocio 1): representa un canal/sala sobre el cual
// los participantes intercambian eventos en tiempo real.
type Room struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// NewRoom valida y construye una Room. La validación vive en el dominio,
// nunca en el adaptador HTTP/WebSocket que la invoca.
func NewRoom(id, name string) (Room, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Room{}, ErrEmptyRoomName
	}
	return Room{ID: id, Name: name, CreatedAt: time.Now().UTC()}, nil
}

// Participant (Entidad de negocio 2): un usuario conectado a una Room.
type Participant struct {
	ID     string
	RoomID string
	Name   string
}

// NewParticipant valida y construye un Participant.
func NewParticipant(id, roomID, name string) (Participant, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Participant{}, ErrEmptyParticipant
	}
	return Participant{ID: id, RoomID: roomID, Name: name}, nil
}

// Event (Entidad de negocio 3): un mensaje/evento emitido por un
// participante dentro de una Room.
type Event struct {
	ID        string
	RoomID    string
	Author    string
	Payload   string
	Timestamp time.Time
}

// NewEvent valida y construye un Event. Estas reglas (no vacío, longitud
// máxima) son negocio puro: deben cumplirse sin importar si el evento
// llegó por WebSocket, por una futura API REST o desde una prueba unitaria.
func NewEvent(id, roomID, author, payload string) (Event, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return Event{}, ErrEmptyPayload
	}
	if len(payload) > maxPayloadLength {
		return Event{}, ErrPayloadTooLong
	}
	return Event{
		ID:        id,
		RoomID:    roomID,
		Author:    author,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}, nil
}
