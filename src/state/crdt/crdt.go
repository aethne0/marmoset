// CRDT types
package crdt

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	pb "marmoset/gen/proto/v1"
	"strconv"
	"strings"

	"github.com/google/btree"
	"github.com/google/uuid"
)

type Tag string

func idstr(id uuid.UUID) string {
	var buf [32]byte
	hex.Encode(buf[:], id[:])
	return string(buf[:])

}
func NewTag(id uuid.UUID, counter uint64) Tag {
	return Tag(fmt.Sprintf("%s%016x", idstr(id), counter))
}

func tagLess(a, b Tag) bool {
	return a < b
}

func newBTree() *btree.BTreeG[Tag] {
	bt := btree.NewG(2, tagLess)
	return bt
}

// OR-Set (set)
// todo these should be sorted for more efficient tag>vvector retrieval

type ORSet struct {
	adds    map[string]*btree.BTreeG[Tag]
	removes *btree.BTreeG[Tag]
}

func NewORSet() *ORSet {
	return &ORSet{
		adds:    make(map[string]*btree.BTreeG[Tag]),
		removes: newBTree(),
	}
}

func (o *ORSet) String() string {
	var s strings.Builder

	s.WriteString("ORSet----------------------------------------------------\n")
	s.WriteString("|--Adds:\n")
	for key, tags := range o.adds {
		s.WriteString(fmt.Sprintf("| |--%s\n", key))
		tags.Ascend(func(tag Tag) bool {
			fmt.Fprintf(&s, "| |  |--%s\n", tag)
			return true
		})
	}

	s.WriteString("|--Removes:\n")
	o.removes.Ascend(func(tag Tag) bool {
		fmt.Fprintf(&s, "| |-----%s\n", tag)
		return true
	})
	s.WriteString("-----------------------------------------------------------")

	return s.String()
}

func (o *ORSet) Add(key string, tag Tag) {
	atags, has := o.adds[key]
	if !has {
		atags = newBTree()
		o.adds[key] = atags
	}

	atags.ReplaceOrInsert(tag)
}

func (o *ORSet) Remove(key string) {
	_, has := o.adds[key]
	if has {
		o.adds[key].Ascend(func(tag Tag) bool {
			o.removes.ReplaceOrInsert(tag)
			return true
		})
	}
}

func (o *ORSet) Contains(key string) bool {
	atags, has := o.adds[key]
	if !has {
		return false
	}

	contains := false
	atags.Ascend(func(tag Tag) bool {
		if !o.removes.Has(tag) {
			contains = true
			return false
		} else {
			return true
		}
	})

	return contains
}

func (o *ORSet) Merge(other *ORSet) {
	for value, otherTags := range other.adds {
		if _, has := o.adds[value]; !has {
			o.adds[value] = newBTree()
		}
		otherTags.Ascend(func(tag Tag) bool {
			o.adds[value].ReplaceOrInsert(tag)
			return true
		})
	}

	other.removes.Ascend(func(tag Tag) bool {
		o.removes.ReplaceOrInsert(tag)
		return true
	})
}

func (o *ORSet) Items() []string {
	var result []string
	for value, tags := range o.adds {
		tags.Ascend(func(tag Tag) bool {
			if !o.removes.Has(tag) {
				result = append(result, value)
				return false
			}
			return true
		})
	}
	return result
}

func ORSetToPB(o *ORSet) *pb.OrSet {
	adds := make([]*pb.ORSetAdd, 0)
	for k, v := range o.adds {
		v.Ascend(func(tag Tag) bool {
			adds = append(adds, &pb.ORSetAdd{Key: k, Tag: string(tag)})
			return true
		})
	}

	removes := make([]*pb.ORSetRemove, 0)
	o.removes.Ascend(func(tag Tag) bool {
		removes = append(removes, &pb.ORSetRemove{Tag: string(tag)})
		return true
	})

	return &pb.OrSet{
		Adds:    adds,
		Removes: removes,
	}
}

func ORSetFromPB(pb *pb.OrSet) *ORSet {
	o := NewORSet()
	for _, t := range pb.Adds {
		o.Add(t.Key, Tag(t.Tag))
	}
	for _, t := range pb.Removes {
		o.removes.ReplaceOrInsert(Tag(t.Tag))
	}
	return o
}

func ORSetToPBDiff(
	o *ORSet,
	peerVector map[uuid.UUID]uint64,
) *pb.OrSet {

	adds := make([]*pb.ORSetAdd, 0)
	for k, v := range o.adds {
		v.Ascend(func(tag Tag) bool {
			// todo we should send these as bytes really maybe
			id := uuid.MustParse(string(tag[:32]))
			ctr, _ := strconv.ParseUint(string(tag[32:]), 16, 64)
			entry, has := peerVector[id]

			if !has || entry < ctr {
				adds = append(adds, &pb.ORSetAdd{Key: k, Tag: string(tag)})
			}

			return true
		})
	}

	removes := make([]*pb.ORSetRemove, 0)
	o.removes.Ascend(func(tag Tag) bool {
		id := uuid.MustParse(string(tag[:32]))
		ctr, _ := strconv.ParseUint(string(tag[32:]), 16, 64)
		entry, has := peerVector[id]

		if !has || entry < ctr {
			removes = append(removes, &pb.ORSetRemove{Tag: string(tag)})
		}

		return true
	})

	slog.Debug("removes", ">", removes)

	return &pb.OrSet{
		Adds:    adds,
		Removes: removes,
	}
}
