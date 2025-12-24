package swim

import (
	"errors"

	"github.com/google/uuid"
)

type Transport interface {
	Send(id uuid.UUID, packet *[]byte) error
	Recv(id uuid.UUID) (*[]byte, error)
}


// For debugging
type PrintUDP struct {}
func (p *PrintUDP) Send(id uuid.UUID, packet *[]byte) error {
	return errors.New("unimplemented")
}
func (p *PrintUDP) Recv(id uuid.UUID) (*[]byte, error) {
	return nil, errors.New("unimplemented")
}


