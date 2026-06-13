package collections

import "testing"

func TestSet_AddAndContains(t *testing.T) {
	set := NewSet[string]()

	if set.Contains("build") {
		t.Fatal("a fresh set should not contain any element")
	}

	set.Add("build")
	if !set.Contains("build") {
		t.Fatal("Contains should report true after Add")
	}
}

func TestSet_AddIsIdempotent(t *testing.T) {
	set := NewSet[string]()

	set.Add("build")
	set.Add("build")

	if set.Len() != 1 {
		t.Fatalf("adding the same element twice should leave Len at 1, got %d", set.Len())
	}
}

func TestSet_Remove(t *testing.T) {
	set := NewSet[string]()
	set.Add("build")

	set.Remove("build")
	if set.Contains("build") {
		t.Fatal("Contains should report false after Remove")
	}
	if set.Len() != 0 {
		t.Fatalf("expected empty set after Remove, got Len %d", set.Len())
	}
}

func TestSet_RemoveAbsentIsNoOp(t *testing.T) {
	set := NewSet[string]()
	set.Add("keep")

	set.Remove("ghost")
	if set.Len() != 1 {
		t.Fatalf("removing an absent element must not change Len, got %d", set.Len())
	}
	if !set.Contains("keep") {
		t.Fatal("removing an absent element must not affect existing members")
	}
}

func TestSet_ToggleRoundTrip(t *testing.T) {
	set := NewSet[string]()

	set.Toggle("task")
	if !set.Contains("task") {
		t.Fatal("Toggle on an absent element should add it")
	}

	set.Toggle("task")
	if set.Contains("task") {
		t.Fatal("Toggle on a present element should remove it")
	}
}

func TestSet_Len(t *testing.T) {
	set := NewSet[int]()
	if set.Len() != 0 {
		t.Fatalf("a fresh set should have Len 0, got %d", set.Len())
	}

	set.Add(1)
	set.Add(2)
	set.Add(3)
	if set.Len() != 3 {
		t.Fatalf("expected Len 3 after three distinct adds, got %d", set.Len())
	}
}

func TestSet_Clear(t *testing.T) {
	set := NewSet[string]()
	set.Add("a")
	set.Add("b")

	set.Clear()
	if set.Len() != 0 {
		t.Fatalf("Clear should empty the set, got Len %d", set.Len())
	}
	if set.Contains("a") {
		t.Fatal("Clear should remove all members")
	}

	// The set must stay usable after Clear.
	set.Add("c")
	if !set.Contains("c") {
		t.Fatal("the set should remain usable after Clear")
	}
}
