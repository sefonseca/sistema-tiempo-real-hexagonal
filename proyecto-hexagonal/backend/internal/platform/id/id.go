// Package id genera identificadores únicos. Se inyecta como función en el
// servicio de aplicación y en el handler, así las pruebas pueden usar IDs
// deterministas.
package id

import (
	"crypto/rand"
	"fmt"
)

// New devuelve un UUID versión 4 (RFC 9562) en su forma canónica.
func New() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand solo falla si el sistema operativo no tiene fuente de
		// entropía; no hay forma razonable de continuar.
		panic(fmt.Sprintf("id: crypto/rand no disponible: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // versión 4
	b[8] = (b[8] & 0x3f) | 0x80 // variante RFC 4122/9562
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
