package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	goredis "github.com/redis/go-redis/v9"

	"contexto-hexagonal/internal/domain"
	"contexto-hexagonal/internal/ports"
)

// EventsChannel es el canal de Redis Pub/Sub compartido por todas las
// instancias del backend (ADR-005).
const EventsChannel = keyPrefix + "events"

// PubSub implementa ports.EventBroadcaster (PUBLISH) y
// ports.EventSubscriber (SUBSCRIBE) sobre un mismo canal de Redis.
type PubSub struct {
	rdb     *goredis.Client
	channel string
}

var (
	_ ports.EventBroadcaster = (*PubSub)(nil)
	_ ports.EventSubscriber  = (*PubSub)(nil)
)

// NewPubSub construye el adaptador de difusión sobre el canal por defecto.
func NewPubSub(rdb *goredis.Client) *PubSub {
	return &PubSub{rdb: rdb, channel: EventsChannel}
}

// Broadcast publica el evento (ya persistido por el servicio de aplicación)
// para que cada instancia lo reenvíe a sus clientes de la sala.
func (p *PubSub) Broadcast(roomID string, e domain.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	rec := toEventRecord(e)
	rec.RoomID = roomID
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("redis: serializando evento para difusión: %w", err)
	}
	if err := p.rdb.Publish(ctx, p.channel, data).Err(); err != nil {
		return fmt.Errorf("redis: publicando evento: %w", err)
	}
	return nil
}

// Subscribe se suscribe al canal y entrega cada evento deserializado. El
// canal devuelto se cierra cuando ctx se cancela.
func (p *PubSub) Subscribe(ctx context.Context) (<-chan domain.Event, error) {
	sub := p.rdb.Subscribe(ctx, p.channel)
	// Receive espera la confirmación de SUBSCRIBE: así un error de conexión
	// se reporta aquí y no se pierde silenciosamente.
	if _, err := sub.Receive(ctx); err != nil {
		_ = sub.Close()
		return nil, fmt.Errorf("redis: suscribiendo a %s: %w", p.channel, err)
	}

	out := make(chan domain.Event, 64)
	go func() {
		defer close(out)
		defer sub.Close()
		msgs := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var rec eventRecord
				if err := json.Unmarshal([]byte(msg.Payload), &rec); err != nil {
					log.Printf("redis: mensaje de pub/sub inválido descartado: %v", err)
					continue
				}
				select {
				case out <- rec.toDomain():
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out, nil
}
