// Package app contiene el servicio de aplicación: la implementación
// concreta del puerto de entrada (ports.EventService). Es el "corazón"
// del hexágono. Solo depende de domain y de las interfaces en ports,
// nunca de un driver concreto (Redis, WebSocket, HTTP, etc.).
package app

import (
	"fmt"

	"contexto-hexagonal/internal/domain"
	"contexto-hexagonal/internal/ports"
)

// eventService implementa ports.EventService.
type eventService struct {
	rooms  ports.RoomRepository
	events ports.EventRepository
	notify ports.EventBroadcaster
	newID  func() string // generador de IDs inyectado (facilita pruebas deterministas)
}

// NewEventService construye el servicio inyectando sus dependencias de
// salida por interfaz. Ni Redis ni WebSocket aparecen en esta firma:
// solo los puertos. Esto es lo que permite cambiar de tecnología sin
// tocar esta función.
func NewEventService(
	rooms ports.RoomRepository,
	events ports.EventRepository,
	notify ports.EventBroadcaster,
	newID func() string,
) ports.EventService {
	return &eventService{rooms: rooms, events: events, notify: notify, newID: newID}
}

func (s *eventService) CreateRoom(id, name string) (domain.Room, error) {
	room, err := domain.NewRoom(id, name)
	if err != nil {
		return domain.Room{}, err
	}
	if err := s.rooms.Save(room); err != nil {
		return domain.Room{}, fmt.Errorf("app: no se pudo guardar la sala: %w", err)
	}
	return room, nil
}

func (s *eventService) Join(participantID, roomID, name string) (domain.Participant, error) {
	if _, found, err := s.rooms.FindByID(roomID); err != nil {
		return domain.Participant{}, fmt.Errorf("app: error consultando la sala: %w", err)
	} else if !found {
		return domain.Participant{}, domain.ErrRoomNotFound
	}
	return domain.NewParticipant(participantID, roomID, name)
}

// PublishEvent es el caso de uso principal end-to-end:
// 1) valida que la sala exista, 2) valida y construye el Event (reglas de
// negocio en domain), 3) lo persiste vía el puerto de salida, y
// 4) lo difunde en tiempo real vía el puerto de broadcaster.
// Cualquier error de infraestructura se envuelve, nunca se ignora.
func (s *eventService) PublishEvent(id, roomID, author, payload string) (domain.Event, error) {
	if _, found, err := s.rooms.FindByID(roomID); err != nil {
		return domain.Event{}, fmt.Errorf("app: error consultando la sala: %w", err)
	} else if !found {
		return domain.Event{}, domain.ErrRoomNotFound
	}

	event, err := domain.NewEvent(id, roomID, author, payload)
	if err != nil {
		return domain.Event{}, err // error de validación de negocio, se propaga tal cual
	}

	if err := s.events.Save(event); err != nil {
		return domain.Event{}, fmt.Errorf("app: no se pudo persistir el evento: %w", err)
	}

	if err := s.notify.Broadcast(roomID, event); err != nil {
		// La persistencia ya se hizo: el evento no se pierde aunque falle
		// la difusión en vivo. Se reporta el error para que el adaptador
		// decida (log, reintento, métrica), pero el dato queda a salvo.
		return event, fmt.Errorf("app: evento guardado pero no se pudo difundir: %w", err)
	}

	return event, nil
}

func (s *eventService) History(roomID string) ([]domain.Event, error) {
	if _, found, err := s.rooms.FindByID(roomID); err != nil {
		return nil, fmt.Errorf("app: error consultando la sala: %w", err)
	} else if !found {
		return nil, domain.ErrRoomNotFound
	}
	return s.events.ListByRoom(roomID)
}
