// Package redis contiene los adaptadores de salida sobre Redis: los
// repositorios de salas y eventos (ADR-003) y el Pub/Sub (ADR-005).
// Implementan interfaces de internal/ports; el dominio nunca importa este
// paquete, la dependencia va siempre del adaptador hacia el puerto.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"contexto-hexagonal/internal/domain"
	"contexto-hexagonal/internal/ports"
)

const (
	keyPrefix      = "hexagonal:"
	roomsIndexKey  = keyPrefix + "rooms" // ZSET: miembro = roomID, score = creación (ms)
	maxEventsPerRm = 1000                // tope de la LIST de eventos por sala (LTRIM)
	opTimeout      = 3 * time.Second     // tiempo máximo por operación contra Redis
)

func roomKey(id string) string       { return keyPrefix + "room:" + id }
func eventsKey(roomID string) string { return keyPrefix + "events:" + roomID }

// NewClient crea un cliente de Redis contra addr (ej. "localhost:6379").
func NewClient(addr string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{Addr: addr})
}

// ---------- Representación persistida (JSON) ----------
// DTOs propios del adaptador: el formato en Redis no depende de cómo estén
// declaradas las entidades del dominio.

type roomRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func toRoomRecord(r domain.Room) roomRecord {
	return roomRecord{ID: r.ID, Name: r.Name, CreatedAt: r.CreatedAt}
}

func (r roomRecord) toDomain() domain.Room {
	return domain.Room{ID: r.ID, Name: r.Name, CreatedAt: r.CreatedAt}
}

type eventRecord struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	Author    string    `json:"author"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

func toEventRecord(e domain.Event) eventRecord {
	return eventRecord{ID: e.ID, RoomID: e.RoomID, Author: e.Author, Payload: e.Payload, Timestamp: e.Timestamp}
}

func (e eventRecord) toDomain() domain.Event {
	return domain.Event{ID: e.ID, RoomID: e.RoomID, Author: e.Author, Payload: e.Payload, Timestamp: e.Timestamp}
}

// ---------- RoomRepository ----------

// Store implementa ports.RoomRepository: cada sala es un STRING JSON bajo
// hexagonal:room:{id} y el índice hexagonal:rooms es un ZSET por fecha.
type Store struct {
	rdb *goredis.Client
}

var _ ports.RoomRepository = (*Store)(nil)

// NewStore construye el repositorio de salas.
func NewStore(rdb *goredis.Client) *Store { return &Store{rdb: rdb} }

func (s *Store) Save(r domain.Room) error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	data, err := json.Marshal(toRoomRecord(r))
	if err != nil {
		return fmt.Errorf("redis: serializando sala: %w", err)
	}
	// MULTI/EXEC: la sala y su entrada en el índice se escriben juntas.
	_, err = s.rdb.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		p.Set(ctx, roomKey(r.ID), data, 0)
		p.ZAdd(ctx, roomsIndexKey, goredis.Z{Score: float64(r.CreatedAt.UnixMilli()), Member: r.ID})
		return nil
	})
	if err != nil {
		return fmt.Errorf("redis: guardando sala: %w", err)
	}
	return nil
}

func (s *Store) FindByID(id string) (domain.Room, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	data, err := s.rdb.Get(ctx, roomKey(id)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return domain.Room{}, false, nil
	}
	if err != nil {
		return domain.Room{}, false, fmt.Errorf("redis: leyendo sala: %w", err)
	}
	var rec roomRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return domain.Room{}, false, fmt.Errorf("redis: deserializando sala: %w", err)
	}
	return rec.toDomain(), true, nil
}

// List devuelve las salas ordenadas de la más reciente a la más antigua.
func (s *Store) List() ([]domain.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	ids, err := s.rdb.ZRevRange(ctx, roomsIndexKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: leyendo índice de salas: %w", err)
	}
	rooms := make([]domain.Room, 0, len(ids))
	if len(ids) == 0 {
		return rooms, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = roomKey(id)
	}
	values, err := s.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: leyendo salas: %w", err)
	}
	for _, v := range values {
		str, ok := v.(string)
		if !ok {
			continue // la key del índice ya no existe: se omite
		}
		var rec roomRecord
		if err := json.Unmarshal([]byte(str), &rec); err != nil {
			return nil, fmt.Errorf("redis: deserializando sala: %w", err)
		}
		rooms = append(rooms, rec.toDomain())
	}
	return rooms, nil
}

// ---------- EventRepository ----------

// EventStore implementa ports.EventRepository: los eventos de cada sala se
// guardan en una LIST hexagonal:events:{roomID}, topada con LTRIM.
type EventStore struct {
	rdb *goredis.Client
}

var _ ports.EventRepository = (*EventStore)(nil)

// NewEventStore construye el repositorio de eventos.
func NewEventStore(rdb *goredis.Client) *EventStore { return &EventStore{rdb: rdb} }

func (s *EventStore) Save(e domain.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	data, err := json.Marshal(toEventRecord(e))
	if err != nil {
		return fmt.Errorf("redis: serializando evento: %w", err)
	}
	key := eventsKey(e.RoomID)
	_, err = s.rdb.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		p.RPush(ctx, key, data)
		p.LTrim(ctx, key, -maxEventsPerRm, -1) // conserva solo los últimos N
		return nil
	})
	if err != nil {
		return fmt.Errorf("redis: guardando evento: %w", err)
	}
	return nil
}

// ListByRoom devuelve los eventos de la sala en orden cronológico.
func (s *EventStore) ListByRoom(roomID string) ([]domain.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	values, err := s.rdb.LRange(ctx, eventsKey(roomID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: leyendo eventos: %w", err)
	}
	events := make([]domain.Event, 0, len(values))
	for _, v := range values {
		var rec eventRecord
		if err := json.Unmarshal([]byte(v), &rec); err != nil {
			return nil, fmt.Errorf("redis: deserializando evento: %w", err)
		}
		events = append(events, rec.toDomain())
	}
	return events, nil
}
