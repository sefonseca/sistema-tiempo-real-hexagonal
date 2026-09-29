import { describe, expect, it } from "vitest";
import { initialState, reducer } from "./state";

const room = { id: "r1", name: "General", createdAt: "2026-01-01T00:00:00Z" };
const event = { id: "e1", roomId: "r1", author: "ana", payload: "hola", timestamp: "2026-01-01T00:00:01Z" };

describe("reducer", () => {
  it("agrega una sala creada al inicio de la lista", () => {
    const s = reducer(initialState, { kind: "server", msg: { type: "room_created", room } });
    expect(s.rooms[0].id).toBe("r1");
  });

  it("guarda el historial al unirse", () => {
    const s = reducer(initialState, {
      kind: "server",
      msg: { type: "joined", participant: { id: "p1", roomId: "r1", name: "ana" }, history: [event] },
    });
    expect(s.events).toHaveLength(1);
    expect(s.participant?.name).toBe("ana");
  });

  it("no duplica eventos repetidos", () => {
    let s = reducer(initialState, { kind: "server", msg: { type: "event", event } });
    s = reducer(s, { kind: "server", msg: { type: "event", event } });
    expect(s.events).toHaveLength(1);
  });

  it("muestra el error del servidor y permite limpiarlo", () => {
    let s = reducer(initialState, { kind: "server", msg: { type: "error", message: "el evento no puede tener contenido vacío" } });
    expect(s.error).toContain("vacío");
    s = reducer(s, { kind: "clearError" });
    expect(s.error).toBeNull();
  });

  it("al salir de la sala limpia participante y eventos", () => {
    let s = reducer(initialState, { kind: "server", msg: { type: "event", event } });
    s = reducer(s, { kind: "leave" });
    expect(s.events).toHaveLength(0);
    expect(s.currentRoom).toBeNull();
  });

  it("si falla al unirse, vuelve al lobby con el error", () => {
    let s = reducer(initialState, { kind: "selectRoom", room });
    s = reducer(s, { kind: "server", msg: { type: "error", message: "el nombre del participante no puede estar vacío" } });
    expect(s.currentRoom).toBeNull();
    expect(s.error).toContain("participante");
  });
});
