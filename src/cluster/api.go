package cluster

import (
	"fmt"
	"log/slog"
	"math"
	"time"
)

func (mgr *ClusterMgr) PrintListPeers() {
	mgr.lock.RLock()

	slog.Debug(fmt.Sprintf("Self | id=%s uri=%s c=%d",
		mgr.Id.String(),
		mgr.uri,
		mgr.Counter,
	))

	slog.Debug("Alive")
	for _, p := range mgr.peers {
		if !p.Dead {
			slog.Debug(fmt.Sprintf("|--Peer | id=%s last_seen=%s uri=%s c=%d f=%d",
				p.Id.String(),
				fmt.Sprintf("%ds ago", int(math.Round(time.Since(p.LastSeen).Seconds()))),
				p.Uri,
				p.Counter,
				p.Failed,
			))
		}
	}

	slog.Debug("Presumed-dead")
	for _, p := range mgr.peers {
		if p.Dead {
			slog.Debug(fmt.Sprintf("|--Peer | id=%s last_seen=%s uri=%s c=%d f=%d",
				p.Id.String(),
				fmt.Sprintf("%ds ago", int(math.Round(time.Since(p.LastSeen).Seconds()))),
				p.Uri,
				p.Counter,
				p.Failed,
			))
		}
	}
	mgr.lock.RUnlock()

}

func (mgr *ClusterMgr) IncCounter() uint64 {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	mgr.Counter++
	return mgr.Counter
}
