package cluster

import (
	"time"

	pb "marmoset/gen/proto/v1"

	"github.com/google/uuid"
)

type Peer struct {
	Id       uuid.UUID
	Uri      string // This is a valid RFC3986 uri, validated by pb
	Counter  uint64
	LastSeen time.Time
}

func PeerFromPB(p *pb.Peer) Peer {
	return Peer{
		Id:       uuid.MustParse(p.Id), // Validated by pb (required+valid)
		Uri:      p.Uri,                // Validated by pb (required+valid)
		Counter:  p.Counter,            // Validated by pb (required)
		LastSeen: p.Lastseen.AsTime(),  // Validated by pb (required)
	}
}
