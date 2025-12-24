package swim

import (
	"net"
	"testing"

	"github.com/google/uuid"
)

func TestACK(t *testing.T) {
	goss := make([]GOSSIP, 0)
	for range GOSSIP_MSG_MAX {
		goss = append(goss, GOSSIP{
			Id:    uuid.New(),
			state: ALIVE,
			IncNo: 77,
		})
	}

	s := ACK{
		SenderId:    uuid.New(),
		SenderIncNo: 5,
		SeqNo:       99,
		Gossip:      goss,
	}

	out, err := s.MarshalJSON()

	if err != nil {
		t.Fatalf("{%+v}", err)
	}

	if len(out) >= 1400 {
		t.Fatalf("message too big: %d / 1400 bytes", len(out))

	}

	// fmt.Println(string(out), strconv.Itoa(len(out)))
}

func TestFWDACK(t *testing.T) {
	goss := make([]GOSSIP, 0)
	for range GOSSIP_MSG_MAX {
		goss = append(goss, GOSSIP{
			Id:    uuid.New(),
			state: ALIVE,
			IncNo: 77,
		})
	}

	s := FWDACK{
		SenderId:      uuid.New(),
		SenderIncNo:   5,
		SeqNo:         99,
		OriginalSeqNo: 5,
		TargetId:      uuid.New(),
		TargetAddr:    net.IPv6zero, // these are 16 bytes
		Gossip:        goss,
	}

	out, err := s.MarshalJSON()

	if err != nil {
		t.Fatalf("{%+v}", err)
	}

	if len(out) >= 1400 {
		t.Fatalf("message too big: %d / 1400 bytes", len(out))

	}

	// fmt.Println(string(out), strconv.Itoa(len(out)))
}
