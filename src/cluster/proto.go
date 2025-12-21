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
	Dead     bool
	Failed   int // failed rpc calls from us, only used for error message throttling
}

func NewPeer(id uuid.UUID, uri string, counter uint64, lastSeen time.Time, dead bool) Peer {
	return Peer{
		Id:       id,
		Uri:      uri,
		Counter:  counter,
		LastSeen: lastSeen,
		Dead:     dead,
		Failed:   0,
	}
}

func NewInitPeer(id uuid.UUID, uri string) Peer {
	return NewPeer(id, uri, 1, time.Now(), false)
}

func PeerFromPB(p *pb.Peer) Peer {
	return NewPeer(
		uuid.MustParse(p.Id), // Validated by pb (required+valid)
		p.Uri,                // Validated by pb (required+valid)
		p.Counter,            // Validated by pb (required)
		p.Lastseen.AsTime(),  // Validated by pb (required)
		p.Dead,
	)
}

func (p *Peer) ToPB() pb.Peer {
	return pb.Peer{
		Id:       p.Id.String(),
		Uri:      p.Uri,
		Counter:  p.Counter,
		Lastseen: timestamppb.New(p.LastSeen),
		Dead:     p.Dead,
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
