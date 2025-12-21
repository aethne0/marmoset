package cluster

import (
	"fmt"
	"log/slog"
	"math"
	"time"
)

func (mgr *ClusterMgr) ListPeers() {
	mgr.lock.RLock()

	slog.Debug("Awake  ------------------------------------")
	for i, p := range mgr.peers {
		if time.Since(p.LastSeen) <= SLEEP_TIMER {
			slog.Debug(fmt.Sprintf("%d: Peer | id=%s last_seen=%s uri=%s c=%d", i, p.Id.String(),
				fmt.Sprintf("%ds ago", int(math.Round(time.Since(p.LastSeen).Seconds()))),
				p.Uri,
				p.Counter))
		}
	}

	slog.Debug("Asleep ------------------------------------")
	for i, p := range mgr.peers {
		if time.Since(p.LastSeen) > SLEEP_TIMER {
			slog.Debug(fmt.Sprintf("%d: zzzz | id=%s last_seen=%s uri=%s c=%d", i, p.Id.String(),
				fmt.Sprintf("%ds ago", int(math.Round(time.Since(p.LastSeen).Seconds()))),
				p.Uri,
				p.Counter))
		}
	}
	mgr.lock.RUnlock()

}
