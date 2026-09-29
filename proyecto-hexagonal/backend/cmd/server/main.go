// Command server es la raíz de composición: el único lugar que conoce a la
// vez el dominio, los puertos y los adaptadores concretos (Redis,
// WebSocket, HTTP) y los ensambla.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	wsadapter "contexto-hexagonal/internal/adapters/in/websocket"
	redisadapter "contexto-hexagonal/internal/adapters/out/redis"
	"contexto-hexagonal/internal/app"
	"contexto-hexagonal/internal/platform/id"
)

func main() {
	port := getenv("PORT", "8080")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---------- Adaptadores de salida ----------
	rdb := redisadapter.NewClient(redisAddr)
	defer rdb.Close()
	if err := waitForRedis(ctx, rdb); err != nil {
		log.Fatalf("redis no disponible en %s: %v", redisAddr, err)
	}
	log.Printf("conectado a redis en %s", redisAddr)

	rooms := redisadapter.NewStore(rdb)
	events := redisadapter.NewEventStore(rdb)
	pubsub := redisadapter.NewPubSub(rdb)

	// ---------- Núcleo (puerto de entrada) ----------
	svc := app.NewEventService(rooms, events, pubsub, id.New)

	// ---------- Adaptadores de entrada ----------
	hub := wsadapter.NewHub()
	go hub.Run(ctx, pubsub)

	mux := http.NewServeMux()
	mux.Handle("/ws", wsadapter.NewHandler(svc, hub, id.New))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(pingCtx).Err(); err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("escuchando en :%s (ws: /ws, health: /healthz)", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("servidor http: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("apagando...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Shutdown no cierra las conexiones "hijacked" por WebSocket: se cierran
	// explícitamente a través del Hub.
	hub.CloseAll()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("apagado forzado: %v", err)
	}
}

// waitForRedis hace PING con backoff exponencial hasta que Redis responda,
// se agoten los intentos o se cancele ctx.
func waitForRedis(ctx context.Context, rdb *goredis.Client) error {
	const attempts = 10
	backoff := 250 * time.Millisecond
	var err error
	for i := 1; i <= attempts; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = rdb.Ping(pingCtx).Err()
		cancel()
		if err == nil {
			return nil
		}
		log.Printf("esperando a redis (intento %d/%d): %v", i, attempts, err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, 5*time.Second)
	}
	return err
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
