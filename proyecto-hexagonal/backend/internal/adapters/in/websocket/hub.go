package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"contexto-hexagonal/internal/domain"
	"contexto-hexagonal/internal/ports"
)

// Hub mantiene el registro de conexiones activas por sala y reenvía a
// cada una los eventos que llegan por el puerto EventSubscriber (ADR-005).
// Como todas las instancias del backend escuchan el mismo canal, un evento
// publicado en la instancia A llega también a los clientes de la instancia B.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*client]struct{}
	all   map[*client]struct{} // todas las conexiones abiertas, estén o no en una sala
}

// NewHub construye un Hub vacío.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*client]struct{}),
		all:   make(map[*client]struct{}),
	}
}

// attach/detach registran la conexión mientras está abierta, para poder
// cerrarla en el apagado aunque todavía no se haya unido a una sala.
func (h *Hub) attach(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.all[c] = struct{}{}
}

func (h *Hub) detach(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.all, c)
}

// Run se suscribe vía el puerto y difunde cada evento a los clientes de su
// sala. Si la suscripción falla o se corta, reintenta con backoff hasta que
// ctx se cancele.
func (h *Hub) Run(ctx context.Context, subscriber ports.EventSubscriber) {
	const (
		minBackoff = 500 * time.Millisecond
		maxBackoff = 10 * time.Second
	)
	backoff := minBackoff
	for ctx.Err() == nil {
		events, err := subscriber.Subscribe(ctx)
		if err != nil {
			log.Printf("hub: no se pudo suscribir (reintento en %s): %v", backoff, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, maxBackoff)
			continue
		}
		backoff = minBackoff
		for e := range events {
			h.BroadcastEvent(e)
		}
	}
}

// Register agrega la conexión a la sala indicada.
func (h *Hub) Register(roomID string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.rooms[roomID]
	if !ok {
		set = make(map[*client]struct{})
		h.rooms[roomID] = set
	}
	set[c] = struct{}{}
}

// Unregister quita la conexión de la sala (no hace nada si no estaba).
func (h *Hub) Unregister(roomID string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.rooms[roomID]
	if !ok {
		return
	}
	delete(set, c)
	if len(set) == 0 {
		delete(h.rooms, roomID)
	}
}

// BroadcastEvent envía el evento a todas las conexiones de su sala,
// incluido el autor.
func (h *Hub) BroadcastEvent(e domain.Event) {
	h.SendToRoom(e.RoomID, newEventMessage(e))
}

// SendToRoom serializa msg una sola vez y lo encola para cada conexión de
// la sala. Una conexión que no da abasto (buffer lleno) se cierra para no
// frenar al resto.
func (h *Hub) SendToRoom(roomID string, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("hub: no se pudo serializar el mensaje: %v", err)
		return
	}

	h.mu.RLock()
	targets := make([]*client, 0, len(h.rooms[roomID]))
	for c := range h.rooms[roomID] {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		if !c.enqueue(data) {
			c.close()
		}
	}
}

// CloseAll cierra todas las conexiones abiertas (apagado ordenado).
func (h *Hub) CloseAll() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.all {
		c.close()
	}
}
