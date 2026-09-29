import type { ChatEvent, Participant, Room } from "./types";

// Mensajes que la aplicación puede recibir del servidor, ya tipados.
export type ServerMessage =
  | { type: "rooms"; rooms: Room[] }
  | { type: "room_created"; room: Room }
  | { type: "joined"; participant: Participant; history: ChatEvent[] }
  | { type: "event"; event: ChatEvent }
  | { type: "error"; message: string };

export type ConnectionState = "connecting" | "open" | "closed";

// PUERTO del frontend: la UI depende de esta interfaz, no de WebSocket.
// El adaptador (adapters/websocketClient.ts) es una de sus implementaciones.
export interface RealtimeGateway {
  connect(): void;
  disconnect(): void;
  listRooms(): void;
  createRoom(name: string): void;
  join(roomId: string, name: string): void;
  publish(payload: string): void;
  onMessage(handler: (msg: ServerMessage) => void): () => void;
  onState(handler: (state: ConnectionState) => void): () => void;
}
