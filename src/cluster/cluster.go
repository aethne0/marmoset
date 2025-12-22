package cluster

import (
	"context"
	"log/slog"
	pb "marmoset/gen/proto/v1"
	pbCon "marmoset/gen/proto/v1/protov1connect"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// How long to stop gossiping to peer if they are unseen - if you have tons of peers this would have to be adjusted
const PEER_TIMOUT time.Duration = time.Duration(15 * time.Second)
const SLEEPER_INTERVAL time.Duration = time.Duration(5 * time.Second)
const GOSSIP_INTERVAL time.Duration = time.Duration(20 * time.Millisecond)

// TODO
// we need to handle the case where we have a node with a stale uri, and a new node comes online with
// that same stale uri. We need to detect "different id than expected" and handle that somehow (sleep old one?).

type ClusterMgr struct {
	Id          uuid.UUID
	Counter     uint64
	lock        sync.RWMutex
	initialized *atomic.Bool
	uri         string
	peers       []Peer
	clients     map[uuid.UUID]pbCon.ClusterServiceClient
}

func NewClusterMgr(uri string, contactUri string) *ClusterMgr {
	initialized := atomic.Bool{}
	if strings.Compare(contactUri, "") == 0 {
		initialized.Store(true)
	}

	mgr := ClusterMgr{
		peers:       make([]Peer, 0),
		clients:     make(map[uuid.UUID]pbCon.ClusterServiceClient, 0),
		Id:          uuid.New(),
		initialized: &initialized,
		uri:         uri,
		Counter:     1, // start at 1 to appease protobuf validation
	}

	slog.Info("Marmoset!", "id", mgr.Id, "uri", uri)

	go mgr.worker(contactUri)
	go mgr.sleeper()

	return &mgr
}

// Must be called with write-lock
func (mgr *ClusterMgr) mergeFromGossipMsg(msg *pb.GossipMsg) {
	g := GossipFromPB(msg)
	g.Peers = append(g.Peers,
		NewPeer(uuid.MustParse(msg.Id), msg.Uri, msg.Counter, time.Now()),
	)
	for _, gossPeer := range g.Peers {
		if gossPeer.Id == mgr.Id {
			continue
		}
		index := slices.IndexFunc(mgr.peers, func(p Peer) bool { return p.Id == gossPeer.Id })
		if index != -1 {
			mgr.peers[index].Counter = max(mgr.peers[index].Counter, gossPeer.Counter)
			if gossPeer.LastSeen.After(mgr.peers[index].LastSeen) {
				mgr.peers[index].LastSeen = gossPeer.LastSeen
			}
		} else {
			mgr.peers = append(mgr.peers, gossPeer)
		}
	}

	// update our counter to highest seen value by either peer
	mgr.Counter = max(mgr.Counter, g.Counter)
}

// Worker stuff

func (mgr *ClusterMgr) worker(contactUri string) {
	for {
		if mgr.initialized.Load() {
			mgr.gossip()
		} else {
			mgr.greet(contactUri)
		}
		time.Sleep(GOSSIP_INTERVAL)
	}
}

// Greet worker
func (mgr *ClusterMgr) greet(contactUri string) {
	if !mgr.initialized.Load() {
		slog.Info("Attempting to contact seed peer", "uri", contactUri)

		// make temporary client with contact uri
		cli := pbCon.NewClusterServiceClient(
			http.DefaultClient,
			contactUri,
		)

		res, err := cli.Greet(
			context.Background(),
			&pb.GreetMsg{Id: mgr.Id.String(), Uri: mgr.uri},
		)
		if err != nil {
			slog.Warn("Couldn't contact seed peer", "err", err)
		} else {
			{
				mgr.lock.Lock()
				mgr.peers = append(
					mgr.peers,
					NewInitPeer(uuid.MustParse(res.Id), res.Uri),
				)
				mgr.initialized.Store(true)
				mgr.lock.Unlock()
			}

			slog.Info(
				"Contacted seed peer successfully, now initialized.",
				"peer-id", res.Id,
				"peer-uri", res.Uri,
			)

			return
		}
	}
}

// ill fix this to work with a wlock eventually? maybe?
func (mgr *ClusterMgr) GetLivePeer() *Peer {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()

	// select random rPeer
	eligable := make([]*Peer, 0)
	for _, p := range mgr.peers {
		if !p.Dead {
			eligable = append(eligable, &p)
		}
	}

	if len(eligable) == 0 {
		return nil
	}

	return eligable[rand.Int()%len(eligable)]
}

// Gossip worker
func (mgr *ClusterMgr) gossip() {

	rPeer := mgr.GetLivePeer() // doesnt matter if its removed by the time we go
	if rPeer == nil {
		return
	}

	mgr.lock.RLock()

	client := mgr.clients[rPeer.Id]
	if client == nil {
		client = pbCon.NewClusterServiceClient(
			http.DefaultClient,
			rPeer.Uri,
		)
		mgr.clients[rPeer.Id] = client
	}

	peers := make([]*pb.Peer, 0)

	for _, peer := range mgr.peers {
		if peer.Id != rPeer.Id {
			pbPeer := peer.ToPB()
			peers = append(peers, &pbPeer)
		}
	}

	msg := &pb.GossipMsg{
		Id:      mgr.Id.String(),
		Uri:     mgr.uri,
		Counter: mgr.Counter,
		Peers:   peers,
	}

	mgr.lock.RUnlock()

	resp, err := client.Gossip(context.Background(), msg)
	if err != nil {
		mgr.lock.Lock()
		index := slices.IndexFunc(mgr.peers, func(p Peer) bool { return p.Id == rPeer.Id })
		var p *Peer
		if index > -1 {
			p = &mgr.peers[index]
			p.Failed++
		}

		if p == nil || p.Failed == 1 {
			slog.Warn("Couldn't gossip to peer", "err", err)
		}
		mgr.lock.Unlock()

		return
	}

	// ? TO SELF - caller managing locks makes locking like this very readable
	// ? im not sure of a way to self-document if a method takes OR needs locks
	// ? rust you can impl Gaurd<T> etc (and usually dont need to because of
	// ? Mutex<T>) )but for Go i dont know
	mgr.lock.Lock()
	mgr.mergeFromGossipMsg(resp)

	if resp.Id != rPeer.Id.String() {
		slog.Warn("Unexpected ID from gossip resp. Presuming dead.", "expected", rPeer.Id.String(), "resp", resp.Id)
		deadIndex := slices.IndexFunc(mgr.peers, func(p Peer) bool { return p.Id == rPeer.Id })
		if deadIndex > -1 {
			mgr.peers[deadIndex].Dead = true
		}
	}
	mgr.lock.Unlock()
}

func (mgr *ClusterMgr) sleeper() {
	// we are checking with an RLock first because its very rare well have to do any writes
	for {
		mgr.lock.Lock()
		for i, p := range mgr.peers {
			if !p.Dead && time.Since(p.LastSeen) > PEER_TIMOUT {
				mgr.peers[i].Dead = true
				slog.Info("Presuming node dead", "id", p.Id)
			}
		}
		mgr.lock.Unlock()

		time.Sleep(SLEEPER_INTERVAL)
	}
}
