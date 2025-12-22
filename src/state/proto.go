package state

import (
	"marmoset/src/state/crdt"

	"github.com/google/uuid"
)

type ReplReq struct {
	Id     uuid.UUID
	Vector map[string]uint64
}

type ReplResp struct {
	Id    uuid.UUID
	OrSet crdt.ORSet
}
