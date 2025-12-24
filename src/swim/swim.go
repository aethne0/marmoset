// This is an abstracted implementation of the swim protocol to handle
// eventual disseminating cluster membership+liveness tracking.
package swim

import "github.com/google/uuid"

type Swimmer struct {
	id    uuid.UUID
	peers []GOSSIP
}

// net <-> channel <-> logic


