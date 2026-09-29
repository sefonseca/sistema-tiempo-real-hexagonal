import type { ConnectionState, RealtimeGateway, ServerMessage } from "../domain/ports";

const DEFAULT_URL = "ws://localhost:8080/ws";

// ADAPTADOR del frontend: implementa RealtimeGateway con WebSocket, con
// reconexión automática (backoff exponencial, máximo 10 s).
export class WebSocketClient implements RealtimeGateway {
  private ws: WebSocket | null = null;
  private messageHandlers = new Set<(m: ServerMessage) => void>();
  private stateHandlers = new Set<(s: ConnectionState) => void>();
  private retry = 0;
  private timer: ReturnType<typeof setTimeout> | null = null;
  private manuallyClosed = false;

  constructor(private readonly url: string = import.meta.env.VITE_WS_URL ?? DEFAULT_URL) {}

  connect(): void {
    this.manuallyClosed = false;
    this.emitState("connecting");
    const ws = new WebSocket(this.url);
    this.ws = ws;

    ws.onopen = () => {
      if (this.ws !== ws) return; // socket obsoleto (p. ej. StrictMode en desarrollo)
      this.retry = 0;
      this.emitState("open");
    };
    ws.onmessage = (ev) => {
      if (this.ws !== ws) return;
      try {
        const msg = JSON.parse(ev.data as string) as ServerMessage;
        this.messageHandlers.forEach((h) => h(msg));
      } catch {
        console.error("Mensaje inválido del servidor", ev.data);
      }
    };
    ws.onclose = () => {
      if (this.ws !== ws) return;
      this.emitState("closed");
      if (!this.manuallyClosed) this.scheduleReconnect();
    };
  }

  disconnect(): void {
    this.manuallyClosed = true;
    if (this.timer) clearTimeout(this.timer);
    this.ws?.close();
  }

  listRooms(): void { this.send({ type: "list_rooms" }); }
  createRoom(name: string): void { this.send({ type: "create_room", name }); }
  join(roomId: string, name: string): void { this.send({ type: "join", roomId, name }); }
  publish(payload: string): void { this.send({ type: "publish", payload }); }

  onMessage(handler: (m: ServerMessage) => void): () => void {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  onState(handler: (s: ConnectionState) => void): () => void {
    this.stateHandlers.add(handler);
    return () => this.stateHandlers.delete(handler);
  }

  private send(data: object): void {
    if (this.ws?.readyState === WebSocket.OPEN) this.ws.send(JSON.stringify(data));
  }

  private emitState(s: ConnectionState): void {
    this.stateHandlers.forEach((h) => h(s));
  }

  private scheduleReconnect(): void {
    const delay = Math.min(1000 * 2 ** this.retry, 10_000);
    this.retry += 1;
    this.timer = setTimeout(() => this.connect(), delay);
  }
}
