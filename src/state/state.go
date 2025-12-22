package state

import (
	"context"
	"fmt"
	"log/slog"
	"marmoset/src/cluster"
	"marmoset/src/state/crdt"
	"net/http"
	"sync"
	"time"

	pb "marmoset/gen/proto/v1"
	pbCon "marmoset/gen/proto/v1/protov1connect"

	"github.com/google/uuid"
)

var PULL_INTERVAL time.Duration = time.Duration(1 * time.Second)

type State struct {
	clusterMgr  *cluster.ClusterMgr
	lock        sync.RWMutex // can be finegrained later
	vector      map[uuid.UUID]uint64
	peerVectors map[uuid.UUID](map[uuid.UUID]uint64) // For each peer, the highest that peer has seen from everyone else
	set         *crdt.ORSet
	clients     map[uuid.UUID]pbCon.StateServiceClient
}

func NewStateMgr(clusterMgr *cluster.ClusterMgr) *State {
	state := State{
		clusterMgr:  clusterMgr,
		lock:        sync.RWMutex{},
		vector:      make(map[uuid.UUID]uint64),
		peerVectors: make(map[uuid.UUID]map[uuid.UUID]uint64),
		set:         crdt.NewORSet(),
		clients:     make(map[uuid.UUID]pbCon.StateServiceClient),
	}

	state.vector[clusterMgr.Id] = 1

	go state.puller()
	return &state
}

func (s *State) setPeerVector(idPeer uuid.UUID, idPeerPeer uuid.UUID, counter uint64) {
	_, has := s.peerVectors[idPeer]
	if !has {
		s.peerVectors[idPeer] = make(map[uuid.UUID]uint64)
	}
	s.peerVectors[idPeer][idPeerPeer] = counter
}

func (s *State) puller() {
	for {
		s.pull()
		time.Sleep(PULL_INTERVAL)
	}
}

func (s *State) pull() {
	p := s.clusterMgr.GetLivePeer()
	if p == nil {
		return
	}

	s.lock.Lock()
	client := s.clients[p.Id]
	if client == nil {
		client = pbCon.NewStateServiceClient(
			http.DefaultClient,
			p.Uri,
		)
		s.clients[p.Id] = client
	}

	vec := make(map[string]uint64)
	for k, v := range s.vector {
		vec[k.String()] = v
	}

	s.lock.Unlock()

	resp, err := client.Replicate(context.Background(),
		&pb.ReplReq{
			Id:     s.clusterMgr.Id.String(),
			Vector: vec,
		},
	)

	if err != nil {
		slog.Warn("Err pulling state", "err", err)
		return
	}

	s.lock.Lock()
	s.set.Merge(crdt.ORSetFromPB(resp.Orset))
	mergeVectors(s.vector, resp.Vector)
	s.lock.Unlock()

}

// replicate endpoint
func (s *State) Replicate(
	_ context.Context,
	req *pb.ReplReq,
) (*pb.ReplResp, error) {
	id := uuid.MustParse(req.Id)

	s.lock.Lock()

	_, has := s.peerVectors[id]
	if !has {
		s.peerVectors[id] = make(map[uuid.UUID]uint64)
	}

	for k, v := range req.Vector {
		s.peerVectors[id][uuid.MustParse(k)] = v
	}

	vec := make(map[string]uint64)
	for k, v := range s.vector {
		vec[k.String()] = v
	}

	resp := &pb.ReplResp{
		Id:     s.clusterMgr.Id.String(),
		Orset:  crdt.ORSetToPBDiff(s.set, s.peerVectors[id]),
		Vector: vec,
	}

	s.lock.Unlock()

	return resp, nil
}

func mergeVectors(dst map[uuid.UUID]uint64, src map[string]uint64) {
	for ks, v := range src {
		k := uuid.MustParse(ks)
		if cur, ok := dst[k]; !ok || v > cur {
			dst[k] = v
		}
	}
}

func (s *State) SetInsert(key string) {
	s.lock.Lock()
	cntr := s.clusterMgr.IncCounter()
	s.set.Add(key, crdt.NewTag(s.clusterMgr.Id, cntr))
	s.vector[s.clusterMgr.Id] = cntr
	s.lock.Unlock()
}

func (s *State) SetRemove(key string) {
	s.lock.Lock()
	cntr := s.clusterMgr.IncCounter()
	s.set.Remove(key)
	s.vector[s.clusterMgr.Id] = cntr
	s.lock.Unlock()
}

func (s *State) SetHas(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.set.Contains(key)
}

func (s *State) PrintORSet() {
	s.lock.RLock()
	fmt.Println(s.set.String())
	s.lock.RUnlock()
}

func (s *State) PrintVector() {
	s.lock.RLock()
	fmt.Printf("Version Vector\n")
	for k, v := range s.vector {
		fmt.Printf("|--%s:%5d\n", k.String(), v)
	}
	s.lock.RUnlock()
}

func (s *State) PrintPeerVectors() {
	s.lock.RLock()
	fmt.Printf("Peer Version Vectors\n")
	for p, pv := range s.peerVectors {
		fmt.Printf("|--Peer %s\n", p.String())
		for k, v := range pv {
			fmt.Printf("|  |--%s:%5d\n", k.String(), v)
		}
	}
	s.lock.RUnlock()
}
