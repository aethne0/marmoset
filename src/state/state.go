package state

import (
	"context"
	"marmoset/src/cluster"
	"marmoset/src/state/crdt"
	"sync"

	pb "marmoset/gen/proto/v1"

	"github.com/google/uuid"
)

type State struct {
	clusterMgr  *cluster.ClusterMgr
	lock        sync.RWMutex                         // can be finegrained later
	peerVectors map[uuid.UUID](map[uuid.UUID]uint64) // For each peer, the highest that peer has seen from everyone else
	set         *crdt.ORSet
}

func NewState(clusterMgr *cluster.ClusterMgr) State {
	return State{
		clusterMgr:  clusterMgr,
		lock:        sync.RWMutex{},
		peerVectors: make(map[uuid.UUID]map[uuid.UUID]uint64),
		set:         crdt.NewOrSet(),
	}
}

func (s *State) Repl(
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

	resp := &pb.ReplResp{
		Id:    s.clusterMgr.Id.String(),
		Orset: ORSetToPB(crdt.NewOrSet()),
	}

	s.lock.Unlock()

	return resp, nil
}

/*
func (s *State) OrSetInsert(key string) {
	tag := crdt.Tag{
		Id:      s.clusterMgr.Id,
		Counter: s.clusterMgr.IncCounter(), // serializable
	}

	s.setL.Lock()
	s.set.Add(key, tag)
	s.setL.Unlock()
}
*/
