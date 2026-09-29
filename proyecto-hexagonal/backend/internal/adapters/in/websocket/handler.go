package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	gws "github.com/gorilla/websocket"

	"contexto-hexagonal/internal/domain"
	"contexto-hexagonal/internal/ports"
)

const (
	writeWait      = 10 * time.Second    // tiempo máximo para escribir un frame
	pongWait       = 60 * time.Second    // sin pong en este lapso, la conexión se da por muerta
	pingPeriod     = (pongWait * 9) / 10 // se envía ping antes de que venza pongWait
	maxMessageSize = 16 * 1024           // bytes; holgado para un payload de 2000 caracteres
	sendBufferSize = 64                  // mensajes pendientes por conexión antes de cerrarla
)

// Errores de negocio del dominio: su mensaje se envía tal cual al cliente.
var domainErrors = []error{
	domain.ErrEmptyRoomName,
	domain.ErrEmptyParticipant,
	domain.ErrEmptyPayload,
	domain.ErrPayloadTooLong,
	domain.ErrRoomNotFound,
}

var (
	errNotJoined   = errors.New("debes unirte a una sala antes de publicar")
	errBadMessage  = errors.New("mensaje inválido: se esperaba JSON con un campo 'type'")
	errUnknownType = errors.New("tipo de mensaje desconocido")
	errInternal    = errors.New("error interno del servidor, intenta de nuevo")
)

// Handler es el adaptador de entrada HTTP→WebSocket. Solo conoce el puerto
// ports.EventService, nunca los adaptadores de salida.
type Handler struct {
	svc      ports.EventService
	hub      *Hub
	newID    func() string
	upgrader gws.Upgrader
}

// NewHandler construye el handler de /ws.
func NewHandler(svc ports.EventService, hub *Hub, newID func() string) *Handler {
	return &Handler{
		svc:   svc,
		hub:   hub,
		newID: newID,
		upgrader: gws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Entorno académico: se acepta cualquier Origin.
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade ya respondió al cliente con el error HTTP correspondiente.
		log.Printf("ws: upgrade fallido: %v", err)
		return
	}
	c := &client{
		conn: conn,
		send: make(chan []byte, sendBufferSize),
		done: make(chan struct{}),
	}
	h.hub.attach(c)
	go c.writePump()
	h.readPump(c)
}

// ---------- Conexión ----------

// client es una conexión WebSocket. gorilla/websocket admite un solo
// lector y un solo escritor concurrentes: readPump lee (y es el único que
// toca roomID/participant) y writePump es el único que escribe.
type client struct {
	conn *gws.Conn
	send chan []byte
	done chan struct{}

	joined      bool
	participant domain.Participant
}

// enqueue intenta dejar el mensaje en el buffer de salida sin bloquear.
func (c *client) enqueue(data []byte) bool {
	select {
	case <-c.done:
		return false
	case c.send <- data:
		return true
	default:
		return false
	}
}

// close cierra la conexión subyacente; readPump detecta el error y limpia.
func (c *client) close() {
	_ = c.conn.Close()
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()
	for {
		select {
		case <-c.done:
			_ = c.conn.WriteControl(gws.CloseMessage,
				gws.FormatCloseMessage(gws.CloseNormalClosure, ""), time.Now().Add(writeWait))
			return
		case data := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(gws.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(gws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Handler) readPump(c *client) {
	defer func() {
		if c.joined {
			h.hub.Unregister(c.participant.RoomID, c)
		}
		h.hub.detach(c)
		close(c.done)
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if gws.IsUnexpectedCloseError(err, gws.CloseGoingAway, gws.CloseNormalClosure, gws.CloseNoStatusReceived) {
				log.Printf("ws: conexión cerrada inesperadamente: %v", err)
			}
			return
		}
		var msg ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil || msg.Type == "" {
			h.reply(c, newErrorMessage(errBadMessage.Error()))
			continue
		}
		h.dispatch(c, msg)
	}
}

// dispatch invoca el caso de uso correspondiente al tipo de mensaje. Ningún
// error aquí cierra la conexión: se le informa al cliente y se sigue leyendo.
func (h *Handler) dispatch(c *client, msg ClientMessage) {
	switch msg.Type {
	case TypeListRooms:
		rooms, err := h.svc.ListRooms()
		if err != nil {
			h.replyError(c, err)
			return
		}
		h.reply(c, newRoomsMessage(rooms))

	case TypeCreateRoom:
		room, err := h.svc.CreateRoom(h.newID(), msg.Name)
		if err != nil {
			h.replyError(c, err)
			return
		}
		h.reply(c, RoomCreatedMessage{Type: TypeRoomCreated, Room: toRoomDTO(room)})

	case TypeJoin:
		h.join(c, msg)

	case TypePublish:
		if !c.joined {
			h.reply(c, newErrorMessage(errNotJoined.Error()))
			return
		}
		p := c.participant
		event, err := h.svc.PublishEvent(h.newID(), p.RoomID, p.Name, msg.Payload)
		if err != nil {
			if event.ID != "" {
				// Persistido pero no difundido: al menos el autor lo ve, y
				// el resto lo recibirá en el historial al (re)unirse.
				log.Printf("ws: %v", err)
				h.reply(c, newEventMessage(event))
				return
			}
			h.replyError(c, err)
		}
		// En el caso normal no se responde aquí: el evento llega a todos
		// (autor incluido) a través del Hub vía Redis Pub/Sub.

	default:
		h.reply(c, newErrorMessage(errUnknownType.Error()+": "+msg.Type))
	}
}

func (h *Handler) join(c *client, msg ClientMessage) {
	participant, err := h.svc.Join(h.newID(), msg.RoomID, msg.Name)
	if err != nil {
		h.replyError(c, err)
		return
	}

	// Una conexión participa en una sola sala a la vez.
	if c.joined {
		h.hub.Unregister(c.participant.RoomID, c)
	}
	// Se registra ANTES de leer el historial: ningún evento publicado entre
	// ambos pasos se pierde (el frontend descarta duplicados por ID).
	h.hub.Register(participant.RoomID, c)
	c.joined = true
	c.participant = participant

	history, err := h.svc.History(participant.RoomID)
	if err != nil {
		h.replyError(c, err)
		return
	}
	h.reply(c, newJoinedMessage(participant, history))
}

// reply encola una respuesta dirigida solo a esta conexión.
func (h *Handler) reply(c *client, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: no se pudo serializar la respuesta: %v", err)
		return
	}
	if !c.enqueue(data) {
		c.close()
	}
}

// replyError envía los errores de dominio tal cual y oculta los de
// infraestructura detrás de un mensaje genérico (se registran en el log).
func (h *Handler) replyError(c *client, err error) {
	for _, de := range domainErrors {
		if errors.Is(err, de) {
			h.reply(c, newErrorMessage(de.Error()))
			return
		}
	}
	log.Printf("ws: error de infraestructura: %v", err)
	h.reply(c, newErrorMessage(errInternal.Error()))
}
