package cluster

import (
	"time"

	pb "marmoset/gen/proto/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (p *Peer) ToPB() pb.Peer {
	return pb.Peer{
		Id:       p.Id.String(),
		Uri:      p.Uri,
		Counter:  p.Counter,
		Lastseen: timestamppb.New(p.LastSeen),
	}
}

type Gossip struct {
	Id      uuid.UUID
	Uri     string // This is a valid RFC3986 uri, validated by pb
	Counter uint64
	Peers   []Peer
}

func GossipFromPB(p *pb.GossipMsg) Gossip {
	peers := make([]Peer, 0, len(p.Peers))
	for _, peer := range p.Peers {
		peers = append(peers, PeerFromPB(peer))
	}

	return Gossip{
		Id:      uuid.MustParse(p.Id),
		Uri:     p.Uri,
		Counter: p.Counter,
		Peers:   peers,
	}
}
