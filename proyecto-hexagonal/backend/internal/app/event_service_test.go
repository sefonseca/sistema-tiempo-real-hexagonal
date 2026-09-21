package app

import (
	"errors"
	"sync"
	"testing"

	"contexto-hexagonal/internal/domain"
)

// ---------- Dobles de prueba de los puertos de salida ----------
// Esto es lo que gana el dominio con Hexagonal: se prueba el caso de uso
// completo sin levantar Redis ni un servidor WebSocket real.

type fakeRoomRepo struct {
	mu    sync.Mutex
	rooms map[string]domain.Room
}

func newFakeRoomRepo() *fakeRoomRepo { return &fakeRoomRepo{rooms: map[string]domain.Room{}} }

func (f *fakeRoomRepo) Save(r domain.Room) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rooms[r.ID] = r
	return nil
}

func (f *fakeRoomRepo) FindByID(id string) (domain.Room, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rooms[id]
	return r, ok, nil
}

type fakeEventRepo struct {
	mu     sync.Mutex
	events map[string][]domain.Event
	failOn string // simula un fallo de infraestructura para probar manejo de errores
}

func newFakeEventRepo() *fakeEventRepo { return &fakeEventRepo{events: map[string][]domain.Event{}} }

func (f *fakeEventRepo) Save(e domain.Event) error {
	if f.failOn == e.RoomID {
		return errors.New("fake: fallo simulado de escritura")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events[e.RoomID] = append(f.events[e.RoomID], e)
	return nil
}

func (f *fakeEventRepo) ListByRoom(roomID string) ([]domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.events[roomID], nil
}

type fakeBroadcaster struct {
	broadcasted []domain.Event
	shouldFail  bool
}

func (f *fakeBroadcaster) Broadcast(roomID string, e domain.Event) error {
	if f.shouldFail {
		return errors.New("fake: fallo simulado de difusión en tiempo real")
	}
	f.broadcasted = append(f.broadcasted, e)
	return nil
}

func idGen(id string) func() string { return func() string { return id } }

// ---------- Casos de prueba ----------

func TestPublishEvent_FlujoPrincipalEndToEnd(t *testing.T) {
	rooms := newFakeRoomRepo()
	events := newFakeEventRepo()
	broadcaster := &fakeBroadcaster{}
	svc := NewEventService(rooms, events, broadcaster, idGen("evt-1"))

	room, err := svc.CreateRoom("room-1", "General")
	if err != nil {
		t.Fatalf("CreateRoom no debería fallar: %v", err)
	}

	event, err := svc.PublishEvent("evt-1", room.ID, "ana", "hola equipo")
	if err != nil {
		t.Fatalf("PublishEvent no debería fallar: %v", err)
	}
	if event.Payload != "hola equipo" {
		t.Errorf("payload esperado 'hola equipo', obtuvo %q", event.Payload)
	}

	history, err := svc.History(room.ID)
	if err != nil {
		t.Fatalf("History no debería fallar: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("se esperaba 1 evento en el histórico, hay %d", len(history))
	}
	if len(broadcaster.broadcasted) != 1 {
		t.Fatalf("se esperaba 1 evento difundido, hay %d", len(broadcaster.broadcasted))
	}
}

func TestPublishEvent_SalaInexistente(t *testing.T) {
	svc := NewEventService(newFakeRoomRepo(), newFakeEventRepo(), &fakeBroadcaster{}, idGen("evt-x"))

	_, err := svc.PublishEvent("evt-x", "sala-que-no-existe", "ana", "hola")
	if !errors.Is(err, domain.ErrRoomNotFound) {
		t.Fatalf("se esperaba domain.ErrRoomNotFound, obtuvo: %v", err)
	}
}

func TestPublishEvent_PayloadVacioEsRechazadoPorElDominio(t *testing.T) {
	rooms := newFakeRoomRepo()
	svc := NewEventService(rooms, newFakeEventRepo(), &fakeBroadcaster{}, idGen("evt-2"))
	room, _ := svc.CreateRoom("room-1", "General")

	_, err := svc.PublishEvent("evt-2", room.ID, "ana", "   ")
	if !errors.Is(err, domain.ErrEmptyPayload) {
		t.Fatalf("se esperaba domain.ErrEmptyPayload, obtuvo: %v", err)
	}
}

func TestPublishEvent_EventoQuedaGuardadoAunqueFalleLaDifusion(t *testing.T) {
	rooms := newFakeRoomRepo()
	events := newFakeEventRepo()
	broadcaster := &fakeBroadcaster{shouldFail: true}
	svc := NewEventService(rooms, events, broadcaster, idGen("evt-3"))
	room, _ := svc.CreateRoom("room-1", "General")

	_, err := svc.PublishEvent("evt-3", room.ID, "ana", "hola")
	if err == nil {
		t.Fatal("se esperaba un error por fallo de difusión")
	}

	history, _ := svc.History(room.ID)
	if len(history) != 1 {
		t.Fatalf("el evento debía quedar persistido pese al fallo de difusión; histórico tiene %d", len(history))
	}
}

func TestJoin_RequiereSalaExistente(t *testing.T) {
	svc := NewEventService(newFakeRoomRepo(), newFakeEventRepo(), &fakeBroadcaster{}, idGen("p-1"))

	_, err := svc.Join("p-1", "sala-fantasma", "ana")
	if !errors.Is(err, domain.ErrRoomNotFound) {
		t.Fatalf("se esperaba domain.ErrRoomNotFound, obtuvo: %v", err)
	}
}
