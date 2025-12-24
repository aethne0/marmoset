package swim

import (
	"encoding/json"
	"net"

	"github.com/google/uuid"
)

// "nothing wrong with a little systems-json"

type STATE uint8

const (
	ALIVE = iota
	SUSPECT
	DEAD
)

func (t STATE) String() string {
	switch t {
	case ALIVE:
		return "ALIVE"
	case SUSPECT:
		return "ALIVE"
	case DEAD:
		return "ALIVE"
	default:
		return "INVALID STATE"
	}
}

type GOSSIP struct {
	Id    uuid.UUID
	state uint8
	IncNo uint64
}

const GOSSIP_MSG_MAX = 16

type PING struct {
	SenderId    uuid.UUID
	SenderIncNo uint64
	SeqNo       uint64
	Gossip      []GOSSIP
}

func (m PING) MarshalJSON() ([]byte, error) {
	type Alias PING
	return json.Marshal(struct {
		Type string `json:"Type"`
		Alias
	}{
		Type:  "PING",
		Alias: Alias(m),
	})
}

type ACK struct {
	SenderId    uuid.UUID
	SenderIncNo uint64
	SeqNo       uint64
	Gossip      []GOSSIP
}

func (m ACK) MarshalJSON() ([]byte, error) {
	type Alias ACK
	return json.Marshal(struct {
		Type string `json:"Type"`
		Alias
	}{
		Type:  "ACK",
		Alias: Alias(m),
	})
}

type PINGREQ struct {
	SenderId    uuid.UUID
	SenderIncNo uint64
	TargetId    uuid.UUID
	TargetAddr  net.IP
	SeqNo       uint64
	Gossip      []GOSSIP
}

func (m PINGREQ) MarshalJSON() ([]byte, error) {
	type Alias PINGREQ
	return json.Marshal(struct {
		Type string `json:"Type"`
		Alias
	}{
		Type:  "PINGREQ",
		Alias: Alias(m),
	})
}

type FWDACK struct {
	OriginalSeqNo uint64
	SenderId      uuid.UUID
	SenderIncNo   uint64
	TargetId      uuid.UUID
	TargetAddr    net.IP
	SeqNo         uint64
	Gossip        []GOSSIP
}

func (m FWDACK) MarshalJSON() ([]byte, error) {
	type Alias FWDACK
	return json.Marshal(struct {
		Type string `json:"Type"`
		Alias
	}{
		Type:  "FWDACK",
		Alias: Alias(m),
	})
}
