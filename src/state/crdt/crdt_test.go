package crdt

import (
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/google/uuid"
)

var IDA uuid.UUID = uuid.MustParse("123ad81d-6aa2-432b-b37b-663b63e8ae82")
var IDB uuid.UUID = uuid.MustParse("abcb52f3-dc77-4e68-9b99-743ab184c899")
var IDC uuid.UUID = uuid.MustParse("fabd383f-717e-4615-af96-d84412f31abb")
var IDD uuid.UUID = uuid.MustParse("def49329-e244-43dd-9b41-b172c9fead61")
var T1 Tag = NewTag(IDA, 1)
var T2 Tag = NewTag(IDA, 2)
var T3 Tag = NewTag(IDA, 3)

func TestORSetEmptyContains(t *testing.T) {
	s := NewOrSet()
	has := s.Contains("wew")
	if has {
		t.Fatalf("Shouldn't contain anything")
	}
}

func TestORSetCommutativity(t *testing.T) {
	// We need 4 sets to test (A + B) vs (B + A) independently
	s1_a := NewOrSet()
	s1_b := NewOrSet()
	s2_a := NewOrSet()
	s2_b := NewOrSet()

	t1 := NewTag(IDA, 1)
	t2 := NewTag(IDB, 2)

	// Set up identical states
	s1_a.Add("x", t1)
	s2_a.Add("x", t1)

	s1_b.Add("x", t2)
	s2_b.Add("x", t2)

	// Path 1: s1_a merge s1_b
	s1_a.Merge(s1_b)
	// Path 2: s2_b merge s2_a
	s2_b.Merge(s2_a)

	items1 := s1_a.Items()
	items2 := s2_b.Items()
	slices.Sort(items1)
	slices.Sort(items2)

	if !slices.Equal(items1, items2) {
		t.Fatalf("commutativity failed: %v vs %v", items1, items2)
	}
}

func TestORSetAddContains(t *testing.T) {
	o := NewOrSet()
	tag := T1
	o.Add("foo", tag)

	if !o.Contains("foo") {
		t.Fatal("expected foo to exist after add")
	}
}

func TestORSetRemove(t *testing.T) {
	o := NewOrSet()
	tag := T1
	o.Add("foo", tag)
	o.Remove("foo")

	if o.Contains("foo") {
		t.Fatal("expected foo to be removed")
	}
}

func TestORSetIdempotentAdd(t *testing.T) {
	o := NewOrSet()
	tag := T1

	o.Add("bar", tag)
	o.Add("bar", tag) // duplicate add

	items := o.Items()
	if len(items) != 1 || items[0] != "bar" {
		t.Fatalf("expected single 'bar', got %v", items)
	}
}

func TestORSetIdempotentRemove(t *testing.T) {
	o := NewOrSet()
	tag := T1
	o.Add("baz", tag)
	o.Remove("baz")
	o.Remove("baz") // duplicate remove

	if o.Contains("baz") {
		t.Fatal("baz should be removed")
	}
}

func TestORSetCommutativity2(t *testing.T) {
	o1 := NewOrSet()
	o2 := NewOrSet()

	t1 := T1
	t2 := T2

	o1.Add("x", t1)
	o2.Add("x", t2)

	// Merge in opposite orders
	o1Copy := o1
	o1.Merge(o2)
	o2.Merge(o1Copy)

	items1 := o1.Items()
	items2 := o2.Items()

	if !slices.Equal(items1, items2) {
		t.Fatalf("commutativity failed: %v vs %v", items1, items2)
	}
}

func TestORSetAssociativity(t *testing.T) {
	// Helper to create a pre-filled set
	setup := func() (*ORSet, *ORSet, *ORSet) {
		a, b, c := NewOrSet(), NewOrSet(), NewOrSet()
		a.Add("a", T1)
		b.Add("b", T2)
		c.Add("c", T3)
		return a, b, c
	}

	// (a ∪ b) ∪ c
	a1, b1, c1 := setup()
	a1.Merge(b1)
	a1.Merge(c1)

	// a ∪ (b ∪ c)
	a2, b2, c2 := setup()
	b2.Merge(c2)
	a2.Merge(b2)

	items1 := a1.Items()
	items2 := a2.Items()
	slices.Sort(items1)
	slices.Sort(items2)

	if !slices.Equal(items1, items2) {
		t.Fatalf("associativity failed: %v vs %v", items1, items2)
	}
}

func TestORSetConcurrentAddRemove(t *testing.T) {
	o1 := NewOrSet()
	o2 := NewOrSet()

	t1 := T1
	t2 := T2

	o1.Add("x", t1)
	o1.Remove("x") // remove before o2 adds

	o2.Add("x", t2)

	o1_merged := o1
	o1_merged.Merge(o2)

	if !o1_merged.Contains("x") {
		t.Fatal("o1 should contain x after merge")
	}

	o2_merged := o2
	o2_merged.Merge(o1)

	if !o2_merged.Contains("x") {
		t.Fatal("o2 should contain x after merge")
	}
}

func TestORSetMultipleTags(t *testing.T) {
	o := NewOrSet()

	t1 := T1
	t2 := T2

	o.Add("val", t1)
	o.Add("val", t2)
	o.Remove("val") // removes both tags?

	// Actually only removes tags observed at time of remove
	// We'll test that removing one tag still preserves others
	o = NewOrSet()
	o.Add("val", t1)
	o.Add("val", t2)
	o.removes.ReplaceOrInsert(t1) // simulate removing only t1

	if !o.Contains("val") {
		t.Fatal("expected val to exist since t2 is not removed")
	}
}

func TestORSetReAdd(t *testing.T) {
	o := NewOrSet()
	t1 := T1
	t2 := T2

	o.Add("a", t1)
	o.Remove("a")
	o.Add("a", t2) // Re-add with new tag

	if !o.Contains("a") {
		t.Fatal("expected 'a' to be present after re-addition")
	}
}

func TestORSetMergeEmpty(t *testing.T) {
	o1 := NewOrSet()
	o1.Add("a", T1)
	o2 := NewOrSet()

	o1.Merge(o2)
	if !o1.Contains("a") || len(o1.Items()) != 1 {
		t.Fatal("merge with empty set should not change content")
	}
}

func TestORSetConcurrentRemovals(t *testing.T) {
	o1 := NewOrSet()
	o2 := NewOrSet()
	tag := T1

	o1.Add("x", tag)
	o2.Merge(o1)

	o1.Remove("x")
	o2.Remove("x")

	o1.Merge(o2)
	if o1.Contains("x") {
		t.Fatal("x should stay removed after merging dual removals")
	}
}

func TestORSetRemoveDoesNotAffectConcurrentAdd(t *testing.T) {
	o1 := NewOrSet()
	o2 := NewOrSet()

	t1 := T1
	t2 := T2

	o1.Add("x", t1)

	// o2 removes "x" WITHOUT seeing t1
	o2.Remove("x")

	// o3 adds "x" with t2
	o1.Add("x", t2)

	o1.Merge(o2)
	if !o1.Contains("x") {
		t.Fatal("x should exist because o2's remove didn't see t1 or t2")
	}
}

func TestORSetMultiItemConvergence(t *testing.T) {
	o1 := NewOrSet()
	o2 := NewOrSet()

	o1.Add("a", T1)
	o2.Add("b", T2)
	o1.Add("c", T3)
	o2.Remove("a") // o2 hasn't seen "a", so this should do nothing

	o1.Merge(o2)
	o2.Merge(o1)

	res1 := o1.Items()
	res2 := o2.Items()
	sort.Strings(res1)
	sort.Strings(res2)

	if !reflect.DeepEqual(res1, res2) {
		t.Errorf("sets did not converge: %v != %v", res1, res2)
	}
}

func TestORSetComplexSequence(t *testing.T) {
	n1, n2, n3 := NewOrSet(), NewOrSet(), NewOrSet()
	nodes := make([]*ORSet, 0)
	nodes = append(nodes, n1)
	nodes = append(nodes, n2)
	nodes = append(nodes, n3)

	// 1. Node 0 adds "data" @ 1
	nodes[0].Add("data", NewTag(IDA, 1))

	// 2. Node 1 adds "data" @ 1 (Concurrent)
	nodes[1].Add("data", NewTag(IDB, 1))

	// 3. Node 0 merges Node 1, then removes "data" (removes both tags)
	nodes[0].Merge(nodes[1])
	nodes[0].Remove("data")

	// 4. Node 2 adds "data" @ 1 (Concurrent to the removal)
	nodes[2].Add("data", NewTag(IDC, 1))

	// 5. Node 1 adds "data" @ 2 (Sequential to its first add)
	nodes[1].Add("data", NewTag(IDB, 2))

	// Final Merge All
	for i := range 3 {
		for j := range 3 {
			nodes[i].Merge(nodes[j])
		}
	}

	if !nodes[0].Contains("data") {
		t.Fatal("Data should exist: Node 2's Tag(2,1) and Node 1's Tag(1,2) were never seen by Node 0's remove")
	}
}

func TestORSetTransitiveMerge(t *testing.T) {
	n1, n2, n3 := NewOrSet(), NewOrSet(), NewOrSet()

	// n1 adds and merges to all
	n1.Add("sync", T1)
	n2.Merge(n1)
	n3.Merge(n2)

	// n1 removes it
	n1.Remove("sync")

	// Propagate: n1 -> n2 -> n3
	n2.Merge(n1)
	n3.Merge(n2)

	if n3.Contains("sync") {
		t.Fatal("n3 should have received the removal transitively from n1 via n2")
	}
}

func TestORSetNetworkPartition(t *testing.T) {
	// Group A
	n1, n2 := NewOrSet(), NewOrSet()
	// Group B
	n3, n4 := NewOrSet(), NewOrSet()

	// --- Phase 1: Partitioned ---
	// Group A adds and removes "apple"
	n1.Add("apple", T1)
	n2.Merge(n1)
	n2.Remove("apple") // n2 sees n1's tag and removes it
	n1.Merge(n2)       // Group A is synced: "apple" is gone

	// Group B independently adds "apple" and "banana"
	n3.Add("apple", T2) // Concurrent add to Group A's remove
	n4.Add("banana", T3)
	n3.Merge(n4)
	n4.Merge(n3) // Group B is synced

	// --- Phase 2: Healing ---
	// Merge n1 (Group A) with n3 (Group B)
	n1.Merge(n3)
	n3.Merge(n1)

	// Propagate to everyone
	n2.Merge(n1)
	n4.Merge(n3)

	// --- Assertions ---
	// "apple" should exist because Group A's remove never saw Group B's tag
	if !n1.Contains("apple") {
		t.Error("apple should exist via Group B's concurrent add")
	}
	if !n1.Contains("banana") {
		t.Error("banana should exist")
	}

	// Convergence
	res1, res2, res3, res4 := n1.Items(), n2.Items(), n3.Items(), n4.Items()
	slices.Sort(res1)
	slices.Sort(res2)
	slices.Sort(res3)
	slices.Sort(res4)
	if !slices.Equal(res1, res3) || !slices.Equal(res1, res4) {
		t.Fatal("Partition healing failed to converge")
	}
}

func TestORSetHighChurnConvergence(t *testing.T) {
	// Assuming NewOrSet returns *ORSet
	n1, n2, n3 := NewOrSet(), NewOrSet(), NewOrSet()

	// --- Step 1: Sequential updates on N1 ---
	n1.Add("key1", NewTag(IDA, 1))
	n1.Add("key1", NewTag(IDA, 2))
	n1.Add("key2", NewTag(IDA, 3))
	// --- Step 2: Concurrent work on N2 & N3 ---
	n2.Add("key1", NewTag(IDB, 1))
	n2.Remove("key1") // N2 removes its Tag{u2, 1}
	n3.Add("key3", NewTag(IDC, 1))
	n3.Add("key1", NewTag(IDC, 2))
	// --- Step 3: Partial Sync ---
	n1.Merge(n2)
	// --- Step 4: More Churn ---
	n1.Add("key2", NewTag(IDA, 4))
	n1.Remove("key2") // Removes {u1, 1} and {u1, 2}
	n2.Add("key4", NewTag(IDB, 2))
	n3.Remove("key3") // Removes {u3, 1}

	// --- Step 5: Global Sync ---
	// Using a slice of pointers to ensure we are modifying the original sets
	allNodes := []*ORSet{n1, n2, n3}
	for range 2 {
		for _, a := range allNodes {
			for _, b := range allNodes {
				a.Merge(b)
			}
		}
	}

	// --- Final State Assertions ---

	// key1 should exist: n1{u1,1, u1,2} and n3{u3,1} survive. n2's tag was removed.
	if !n1.Contains("key1") {
		t.Error("key1 should exist")
	}

	// key2 should be gone: N1's remove saw both of its own tags.
	if n1.Contains("key2") {
		t.Error("key2 should be removed")
	}

	// key3 should be gone: N3 removed its own tag.
	if n1.Contains("key3") {
		t.Error("key3 should be removed")
	}

	// key4 should exist: Added by N2, never removed.
	if !n1.Contains("key4") {
		t.Error("key4 should exist")
	}

	// Convergence Check
	res1, res2, res3 := n1.Items(), n2.Items(), n3.Items()
	slices.Sort(res1)
	slices.Sort(res2)
	slices.Sort(res3)

	if !slices.Equal(res1, res2) || !slices.Equal(res2, res3) {
		t.Fatalf("Nodes did not converge!\nN1: %v\nN2: %v\nN3: %v", res1, res2, res3)
	}
}
