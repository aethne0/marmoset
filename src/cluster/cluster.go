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
const SLEEP_TIMER time.Duration = time.Duration(15 * time.Second)
const GOSSIP_INTERVAL time.Duration = time.Duration(1 * time.Second)

// TODO
// we need to handle the case where we have a node with a stale uri, and a new node comes online with
// that same stale uri. We need to detect "different id than expected" and handle that somehow (sleep old one?).

type ClusterMgr struct {
	lock    sync.RWMutex
	peers   []Peer
	clients map[uuid.UUID]pbCon.ClusterServiceClient

	id          uuid.UUID
	initialized *atomic.Bool
	uri         string
}

func NewClusterMgr(uri string, contactUri string) *ClusterMgr {
	initialized := atomic.Bool{}
	if strings.Compare(contactUri, "") == 0 {
		initialized.Store(true)
	}

	mgr := ClusterMgr{
		peers:       make([]Peer, 0),
		clients:     make(map[uuid.UUID]pbCon.ClusterServiceClient, 0),
		id:          uuid.New(),
		initialized: &initialized,
		uri:         uri,
	}

	slog.Info("Marmoset!", "id", mgr.id, "uri", uri)

	go mgr.worker(contactUri)

	return &mgr
}

// must be called with lock
func (mgr *ClusterMgr) mergeFromGossipMsg(msg *pb.GossipMsg) {
	g := GossipFromPB(msg)
	g.Peers = append(g.Peers,
		Peer{Id: uuid.MustParse(msg.Id), Uri: msg.Uri, Counter: msg.Counter, LastSeen: time.Now()},
	)
	for _, gossPeer := range g.Peers {
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
			&pb.GreetMsg{Id: mgr.id.String(), Uri: mgr.uri},
		)
		if err != nil {
			slog.Warn("Couldn't contact seed peer", "err", err)
		} else {
			{
				mgr.lock.Lock()
				mgr.peers = append(
					mgr.peers,
					Peer{
						Id:       uuid.MustParse(res.Id),
						Uri:      res.Uri,
						Counter:  1,
						LastSeen: time.Now(),
					},
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

// Gossip worker
func (mgr *ClusterMgr) gossip() {
	mgr.lock.RLock()

	if len(mgr.peers) == 0 {
		mgr.lock.RUnlock()
		return
	}

	// select random rPeer
	eligable := make([]*Peer, 0)
	for _, p := range mgr.peers {
		if time.Since(p.LastSeen) <= SLEEP_TIMER {
			eligable = append(eligable, &p)
		}
	}
	rPeer := eligable[rand.Int()%len(eligable)]

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
		Id:      mgr.id.String(),
		Uri:     mgr.uri,
		Counter: 1, // todo
		Peers:   peers,
	}

	mgr.lock.RUnlock()

	resp, err := client.Gossip(context.Background(), msg)
	if err != nil {
		slog.Warn("Couldn't gossip to peer", "err", err)
		return
	}

	// ? TO SELF - caller managing locks makes locking like this very readable
	// ? im not sure of a way to self-document if a method takes OR needs locks
	// ? rust you can impl Gaurd<T> etc (and usually dont need to because of
	// ? Mutex<T>) )but for Go i dont know
	mgr.lock.Lock()
	mgr.mergeFromGossipMsg(resp)
	mgr.lock.Unlock()
}
