import { useEffect, useRef, useState } from "react";
import type { ChatEvent, Participant, Room } from "../domain/types";

interface Props {
  room: Room;
  me: Participant | null;
  events: ChatEvent[];
  onSend: (text: string) => void;
  onLeave: () => void;
}

export function ChatRoom({ room, me, events, onSend, onLeave }: Props) {
  const [text, setText] = useState("");
  const bottom = useRef<HTMLDivElement>(null);

  useEffect(() => { bottom.current?.scrollIntoView({ behavior: "smooth" }); }, [events]);

  const submit = () => {
    onSend(text);
    setText("");
  };

  return (
    <section className="card chat">
      <header className="row between">
        <h2>Sala: {room.name}</h2>
        <button className="secondary" onClick={onLeave}>Salir</button>
      </header>

      <div className="messages" aria-live="polite">
        {events.length === 0 && <p className="muted">Aún no hay mensajes.</p>}
        {events.map((e) => (
          <div key={e.id} className={`msg ${me && e.author === me.name ? "mine" : ""}`}>
            <strong>{e.author}</strong>
            <span>{e.payload}</span>
            <time>{new Date(e.timestamp).toLocaleTimeString()}</time>
          </div>
        ))}
        <div ref={bottom} />
      </div>

      <div className="row">
        <input
          placeholder="Escribe un mensaje…"
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && submit()}
        />
        <button onClick={submit}>Enviar</button>
      </div>
    </section>
  );
}
