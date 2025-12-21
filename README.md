# marmoset

I am learning Go in the most normal way to learn a programming language: write a complex distributed storage system from scratch

**marmoset** is a distributed in-memory key-value database. It has the following properties:

- Supports shared sets, maps, counters and sequences
- Eventually-consistent (nodes can have stale state)
- In-memory (if all nodes die data will be lost permanently)
- CRDT based
- Gossip-based membership

<p align="center">
  <img src="https://monke.ca/assets/marmo_set.webp" alt="Marmo-Set"/>
</p>

## design notes (mostly for myself)

### gossip

A node will periodically "gossip" to a randomly selected peer that it knows of. This is a
symmetrical operation - both nodes send/recv a `GossipMsg` which contains (importantly) the
ID and Counter for the selected peer, as well as all peers that peer knows of (one will usually
have more up to date information on one or more peers).

To maintain a sort of lamport-clock we\* also set our counter to `max(our_counter, their_counter)`, under
the assumption that ALSO the other peer has set their counter to `max(their_counter, max(their_peers_counters))`
before constructing+sending the GossipMsg. We will have done the same, and so the `max(our_counter, their_counter)`
will give us the correct highest-seen-counter by either peer.

_(we as in: either node in this rpc exchange, again this is a symmetrical RPC call `GossipMsg->GossipMsg`)_

### crdt GC

GC relies on some notion of "have all nodes seen up to a certain point" (which well then notionally GC
up to). "All nodes" is the tricky part here, because nodes can be offline for 2 days then randomly reappear
asking "yo what i miss".

The tentative plan is to have some state that a known-peer can be in that is "presumed-dead" basically, and even
if that node comes back we will tell it "we thought you were dead so you are going to have to undergo mandatory
reeducation". I believe this will be sound - the risk potentially is if some nodes think a peer is dead but others
dont, but because we require some kind of consensus (in terms of up to dateness) already to GC something (again,
checking that "all nodes" have witnessed a certain timestamp/tag) then we can by the same logic make sure "all nodes"
have recognized that a certain other node is "presumed dead".

the `alive -> presumed_dead` transition should be triggered by a timeout, but we should be much more careful about
the `presumed_dead -> alive` process, which is a bit more dangerous. It would be sub-optimal but safe to tell any
presumed dead node "sorry, but you're dead. Wipe your data and generate a new ID and come back".
