package crdt

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type Tag struct {
	Id      uuid.UUID
	Counter uint64
}

// OR-Set (set)

type ORSet struct {
	adds    map[string]*Set[Tag]
	removes *Set[Tag]
}

func NewOrSet() *ORSet {
	return &ORSet{
		adds:    map[string]*Set[Tag]{},
		removes: NewSet[Tag](),
	}
}

func cmpSets(a Tag, b Tag) int {
	if a.Id == b.Id && a.Counter == b.Counter {
		return 0
	}
	if a.Id == b.Id {
		if a.Counter < b.Counter {
			return -1
		} else {
			return 1
		}
	}
	return strings.Compare(a.Id.String(), b.Id.String())
}

func (o *ORSet) String() string {
	var s strings.Builder

	s.WriteString("ORSet----------------------------------------------------\n")
	s.WriteString("|--Adds:\n")
	for k, v := range o.adds {
		s.WriteString(fmt.Sprintf("| |--%s\n", k))
		vv := v.Entries()
		slices.SortFunc(vv, cmpSets)
		for _, vvv := range vv {
			s.WriteString(fmt.Sprintf("| |  |--%s:%d\n", strings.Split(vvv.Id.String(), "-")[0], vvv.Counter))
		}
	}
	s.WriteString("|--Removes:\n")
	v := o.removes.Entries()
	slices.SortFunc(v, cmpSets)
	for _, v := range v {
		s.WriteString(fmt.Sprintf("| |--%s:%d\n", strings.Split(v.Id.String(), "-")[0], v.Counter))
	}

	s.WriteString("-----------------------------------------------------------")

	return s.String()
}

func (o *ORSet) Add(value string, tag Tag) {
	s, has := o.adds[value]
	if !has {
		s = NewSet[Tag]()
		o.adds[value] = s
	}

	s.Add(tag)
}

func (o *ORSet) Remove(value string) {
	_, has := o.adds[value]
	if has {
		for _, tag := range o.adds[value].Entries() {
			o.removes.Add(tag)
		}
	}
}

func (o *ORSet) Contains(value string) bool {
	for _, tag := range o.adds[value].Entries() {
		if removed := o.removes.Has(tag); !removed {
			return true
		}
	}
	return false
}

func (o *ORSet) Merge(other *ORSet) {
	for value, otherTags := range other.adds {
		if _, has := o.adds[value]; !has {
			o.adds[value] = NewSet[Tag]()
		}
		for _, tag := range otherTags.Entries() {
			o.adds[value].Add(tag)
		}
	}

	for _, tag := range other.removes.Entries() {
		o.removes.Add(tag)
	}
}

func (o *ORSet) Items() []string {
	var result []string
	for value, tags := range o.adds {
		for _, tag := range tags.Entries() {
			if !o.removes.Has(tag) {
				result = append(result, value)
				break
			}
		}
	}
	return result
}
