# marmoset

**marmoset** is a distributed in-memory key-value database. It has the following properties:

- Supports shared sets, maps, counters and sequences
- Eventually-consistent (nodes can have stale state)
- In-memory (if all nodes die data will be lost permanently)
- CRDT based
- Gossip-based membership

<p align="center">
  <img src="https://monke.ca/assets/marmo_set.webp" alt="Marmo-Set"/>
</p>
