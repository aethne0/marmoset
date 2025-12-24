// CRDT types
package crdt

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/btree"
	"github.com/google/uuid"
)

type Tag string

func uuidstr(id uuid.UUID) string {
	var buf [32]byte
	hex.Encode(buf[:], id[:])
	return string(buf[:])

}
func NewTag(id uuid.UUID, counter uint64) Tag {
	return Tag(fmt.Sprintf("%s%016x", uuidstr(id), counter))
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
