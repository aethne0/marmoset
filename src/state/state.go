package state

import "marmoset/src/cluster"

type State struct {
	clusterMgr *cluster.ClusterMgr
}

func NewState(clusterMgr *cluster.ClusterMgr) State {
	return State{clusterMgr: clusterMgr}
}
