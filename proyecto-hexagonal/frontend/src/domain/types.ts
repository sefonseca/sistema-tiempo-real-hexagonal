// Modelo del frontend: independiente del transporte (WebSocket, REST, etc.).

export interface Room {
  id: string;
  name: string;
  createdAt: string;
}

export interface Participant {
  id: string;
  roomId: string;
  name: string;
}

export interface ChatEvent {
  id: string;
  roomId: string;
  author: string;
  payload: string;
  timestamp: string;
}
