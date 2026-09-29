import { useEffect, useMemo, useReducer } from "react";
import { WebSocketClient } from "./adapters/websocketClient";
import type { RealtimeGateway } from "./domain/ports";
import { initialState, reducer } from "./domain/state";
import { Lobby } from "./components/Lobby";
import { ChatRoom } from "./components/ChatRoom";

const statusLabel = { connecting: "Conectando…", open: "Conectado", closed: "Desconectado" } as const;

export default function App() {
  // La UI depende del PUERTO (RealtimeGateway); aquí se inyecta el adaptador WebSocket.
  const gateway: RealtimeGateway = useMemo(() => new WebSocketClient(), []);
  const [state, dispatch] = useReducer(reducer, initialState);

  useEffect(() => {
    const offMsg = gateway.onMessage((msg) => dispatch({ kind: "server", msg }));
    const offState = gateway.onState((s) => {
      dispatch({ kind: "connection", state: s });
      if (s === "open") gateway.listRooms();
    });
    gateway.connect();
    return () => { offMsg(); offState(); gateway.disconnect(); };
  }, [gateway]);

  const online = state.connection === "open";

  return (
    <main>
      <header className="top">
        <h1>Salas en Tiempo Real</h1>
        <span className={`badge ${state.connection}`}>{statusLabel[state.connection]}</span>
      </header>

      {state.error && (
        <div className="error" role="alert" onClick={() => dispatch({ kind: "clearError" })}>
          ⚠ {state.error} <small>(clic para cerrar)</small>
        </div>
      )}

      {state.currentRoom ? (
        <ChatRoom
          room={state.currentRoom}
          me={state.participant}
          events={state.events}
          onSend={(t) => gateway.publish(t)}
          onLeave={() => { dispatch({ kind: "leave" }); gateway.listRooms(); }}
        />
      ) : (
        <Lobby
          rooms={state.rooms}
          disabled={!online}
          onCreate={(name) => gateway.createRoom(name)}
          onJoin={(room, user) => { dispatch({ kind: "selectRoom", room }); gateway.join(room.id, user); }}
        />
      )}
    </main>
  );
}
