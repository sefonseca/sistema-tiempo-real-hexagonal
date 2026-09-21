package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewRoom_RechazaNombreVacio(t *testing.T) {
	_, err := NewRoom("r-1", "   ")
	if !errors.Is(err, ErrEmptyRoomName) {
		t.Fatalf("se esperaba ErrEmptyRoomName, obtuvo: %v", err)
	}
}

func TestNewParticipant_RechazaNombreVacio(t *testing.T) {
	_, err := NewParticipant("p-1", "r-1", "")
	if !errors.Is(err, ErrEmptyParticipant) {
		t.Fatalf("se esperaba ErrEmptyParticipant, obtuvo: %v", err)
	}
}

func TestNewEvent_RechazaPayloadVacio(t *testing.T) {
	_, err := NewEvent("e-1", "r-1", "ana", "")
	if !errors.Is(err, ErrEmptyPayload) {
		t.Fatalf("se esperaba ErrEmptyPayload, obtuvo: %v", err)
	}
}

func TestNewEvent_RechazaPayloadDemasiadoLargo(t *testing.T) {
	huge := strings.Repeat("a", maxPayloadLength+1)
	_, err := NewEvent("e-1", "r-1", "ana", huge)
	if !errors.Is(err, ErrPayloadTooLong) {
		t.Fatalf("se esperaba ErrPayloadTooLong, obtuvo: %v", err)
	}
}

func TestNewEvent_Valido(t *testing.T) {
	e, err := NewEvent("e-1", "r-1", "ana", "hola equipo")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if e.Payload != "hola equipo" || e.Author != "ana" {
		t.Fatalf("evento construido incorrectamente: %+v", e)
	}
}
