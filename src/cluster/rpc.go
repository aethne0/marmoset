package cluster

import (
	"context"
	"fmt"
	"log/slog"
	pb "marmoset/gen/proto/v1"
	"marmoset/src/assert"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

func (mgr *ClusterMgr) Greet(
	_ context.Context,
	req *pb.GreetMsg,
) (*pb.GreetMsg, error) {
	id := uuid.MustParse(req.Id)

	mgr.lock.Lock() // just take write lock, we almost always will do the append
	defer mgr.lock.Unlock()

	if slices.ContainsFunc(mgr.peers, func(p Peer) bool { return p.Id == id }) {
		slog.Warn("Duplicate peer tried to join cluster", "peer-id", id)
		return nil, connect.NewError(
			connect.CodeAlreadyExists,
			fmt.Errorf("Peer already member with requested ID - %s", id.String()),
		)
	}

	mgr.peers = append(
		mgr.peers,
		Peer{
			Id:       id,
			Uri:      req.Uri,
			Counter:  0,
			LastSeen: time.Now(),
		},
	)

	slog.Info("Greeted by new peer", "peer-id", id.String(), "peer-uri", req.Uri)

	return &pb.GreetMsg{
		Id:  mgr.id.String(),
		Uri: mgr.uri,
	}, nil
}

func (mgr *ClusterMgr) Gossip(
	_ context.Context,
	req *pb.GossipMsg,
) (*pb.GossipMsg, error) {
	assert.Todo()
	return nil, nil
}
