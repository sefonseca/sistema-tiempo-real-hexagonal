// Package integration prueba el flujo completo contra un Redis real, con
// los adaptadores de salida concretos (no fakes). Se salta si REDIS_ADDR no
// está definida:
//
//	REDIS_ADDR=localhost:6379 go test ./internal/integration -v
package integration

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	redisadapter "contexto-hexagonal/internal/adapters/out/redis"
	"contexto-hexagonal/internal/app"
	"contexto-hexagonal/internal/domain"
	"contexto-hexagonal/internal/platform/id"
)

func TestE2E_CreateRoomJoinPublishHistory(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR no definida: se omite la prueba de integración")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rdb := redisadapter.NewClient(addr)
	// t.Cleanup (y no defer): las limpiezas corren en orden inverso, así el
	// cliente se cierra DESPUÉS de borrar las keys de la prueba.
	t.Cleanup(func() { _ = rdb.Close() })
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("no se pudo conectar a Redis en %s: %v", addr, err)
	}

	pubsub := redisadapter.NewPubSub(rdb)
	svc := app.NewEventService(
		redisadapter.NewStore(rdb),
		redisadapter.NewEventStore(rdb),
		pubsub,
		id.New,
	)

	// Suscripción real antes de publicar, para verificar la difusión.
	subCtx, subCancel := context.WithCancel(ctx)
	defer subCancel()
	incoming, err := pubsub.Subscribe(subCtx)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	roomID := "e2e-" + id.New()
	t.Cleanup(func() {
		cctx, ccancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ccancel()
		rdb.Del(cctx, "hexagonal:room:"+roomID, "hexagonal:events:"+roomID)
		rdb.ZRem(cctx, "hexagonal:rooms", roomID)
	})

	// 1) CreateRoom
	room, err := svc.CreateRoom(roomID, "Sala E2E")
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	rooms, err := svc.ListRooms()
	if err != nil {
		t.Fatalf("ListRooms: %v", err)
	}
	if !containsRoom(rooms, room.ID) {
		t.Fatalf("ListRooms no incluye la sala recién creada %s", room.ID)
	}

	// 2) Join
	participant, err := svc.Join(id.New(), room.ID, "ana")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	if _, err := svc.Join(id.New(), "sala-inexistente-"+id.New(), "ana"); !errors.Is(err, domain.ErrRoomNotFound) {
		t.Fatalf("Join a sala inexistente: se esperaba ErrRoomNotFound, obtuvo %v", err)
	}

	// 3) PublishEvent
	published, err := svc.PublishEvent(id.New(), room.ID, participant.Name, "hola desde la prueba e2e")
	if err != nil {
		t.Fatalf("PublishEvent: %v", err)
	}

	select {
	case got := <-incoming:
		if got.ID != published.ID || got.RoomID != room.ID || got.Payload != published.Payload {
			t.Fatalf("evento difundido distinto al publicado: %+v vs %+v", got, published)
		}
	case <-ctx.Done():
		t.Fatal("no llegó el evento por Redis Pub/Sub")
	}

	// 4) History
	history, err := svc.History(room.ID)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("se esperaba 1 evento en el histórico, hay %d", len(history))
	}
	if h := history[0]; h.ID != published.ID || h.Author != "ana" || h.Payload != published.Payload {
		t.Fatalf("evento del histórico no coincide: %+v", h)
	}
}

func containsRoom(rooms []domain.Room, id string) bool {
	for _, r := range rooms {
		if r.ID == id {
			return true
		}
	}
	return false
}
