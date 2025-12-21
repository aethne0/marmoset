package state

import (
	pb "marmoset/gen/proto/v1"
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

func ORSetToPB(o *crdt.ORSet) *pb.OrSet {
	return &pb.OrSet{}
}
