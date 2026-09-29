import { useState } from "react";
import type { Room } from "../domain/types";

interface Props {
  rooms: Room[];
  disabled: boolean;
  onCreate: (roomName: string) => void;
  onJoin: (room: Room, userName: string) => void;
}

export function Lobby({ rooms, disabled, onCreate, onJoin }: Props) {
  const [userName, setUserName] = useState("");
  const [roomName, setRoomName] = useState("");

  return (
    <section className="card">
      <h2>1. Tu nombre</h2>
      <input
        placeholder="Ej. Ana"
        value={userName}
        onChange={(e) => setUserName(e.target.value)}
        maxLength={30}
      />

      <h2>2. Crea una sala</h2>
      <div className="row">
        <input
          placeholder="Nombre de la sala"
          value={roomName}
          onChange={(e) => setRoomName(e.target.value)}
          maxLength={40}
        />
        <button
          disabled={disabled}
          onClick={() => { onCreate(roomName); setRoomName(""); }}
        >
          Crear
        </button>
      </div>

      <h2>3. Únete a una sala</h2>
      {rooms.length === 0 && <p className="muted">Aún no hay salas. Crea la primera.</p>}
      <ul className="rooms">
        {rooms.map((r) => (
          <li key={r.id}>
            <span>{r.name}</span>
            <button disabled={disabled} onClick={() => onJoin(r, userName)}>Entrar</button>
          </li>
        ))}
      </ul>
    </section>
  );
}
