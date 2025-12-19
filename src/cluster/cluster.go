package cluster

import (
	"context"
	"log/slog"
	pb "marmoset/gen/proto/v1"
	pbCon "marmoset/gen/proto/v1/protov1connect"
	"marmoset/src/assert"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const GOSSIP_INTERVAL time.Duration = time.Duration(1000 * time.Millisecond)

type ClusterMgr struct {
	lock    sync.RWMutex
	peers   []Peer
	clients map[uuid.UUID]int // todo

	id          uuid.UUID
	initialized *atomic.Bool
	uri         string
}

func NewClusterMgr(uri string, contact string) *ClusterMgr {
	initialized := atomic.Bool{}
	if strings.Compare(contact, "") == 0 {
		initialized.Store(true)
	}

	mgr := ClusterMgr{
		peers:       make([]Peer, 0),
		clients:     make(map[uuid.UUID]int, 0), // todo
		id:          uuid.New(),
		initialized: &initialized,
		uri:         uri,
	}

	if !initialized.Load() {
		go mgr.greet(contact)
	}

	go mgr.gossip()

	return &mgr
}

func (mgr *ClusterMgr) greet(contact string) {
	for {
		if !mgr.initialized.Load() {
			slog.Info("Attempting to contact seed peer", "uri", mgr.uri)

			// make temporary client with contact uri
			cli := pbCon.NewClusterServiceClient(
				http.DefaultClient,
				contact,
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
							Counter:  0,
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
			}
		}
		time.Sleep(GOSSIP_INTERVAL)
	}
}

func (mgr *ClusterMgr) gossip() {
	for {
		if mgr.initialized.Load() {
			assert.Todo()
		}
		time.Sleep(GOSSIP_INTERVAL)
	}
}
