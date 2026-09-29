import type { ServerMessage, ConnectionState } from "./ports";
import type { ChatEvent, Participant, Room } from "./types";

export interface AppState {
  connection: ConnectionState;
  rooms: Room[];
  participant: Participant | null;
  currentRoom: Room | null;
  events: ChatEvent[];
  error: string | null;
}

export type Action =
  | { kind: "connection"; state: ConnectionState }
  | { kind: "server"; msg: ServerMessage }
  | { kind: "selectRoom"; room: Room }
  | { kind: "leave" }
  | { kind: "clearError" };

export const initialState: AppState = {
  connection: "connecting",
  rooms: [],
  participant: null,
  currentRoom: null,
  events: [],
  error: null,
};

// Reductor puro: fácil de probar sin navegador ni WebSocket.
export function reducer(state: AppState, action: Action): AppState {
  switch (action.kind) {
    case "connection":
      return { ...state, connection: action.state };
    case "selectRoom":
      return { ...state, currentRoom: action.room, error: null };
    case "leave":
      return { ...state, currentRoom: null, participant: null, events: [] };
    case "clearError":
      return { ...state, error: null };
    case "server": {
      const m = action.msg;
      switch (m.type) {
        case "rooms":
          return { ...state, rooms: m.rooms };
        case "room_created":
          return { ...state, rooms: [m.room, ...state.rooms.filter((r) => r.id !== m.room.id)] };
        case "joined":
          return { ...state, participant: m.participant, events: m.history ?? [], error: null };
        case "event":
          // evita duplicados si el mismo evento llega dos veces
          return state.events.some((e) => e.id === m.event.id)
            ? state
            : { ...state, events: [...state.events, m.event] };
        case "error":
          // si falla el "join" (participante aún sin confirmar), se vuelve al lobby
          return state.currentRoom && !state.participant
            ? { ...state, error: m.message, currentRoom: null }
            : { ...state, error: m.message };
      }
    }
  }
}
