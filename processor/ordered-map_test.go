package processor

import "testing"

func prepMap() *OrderedMap {
	m := NewOrderedMap()
	m.Add("1", &QPayload{"1111"})
	m.Add("2", &QPayload{"2222"})
	m.Add("3", &QPayload{"3333"})

	return m
}

func TestOrderedMap_Add(t *testing.T) {
	m := prepMap()

	keys := m.GetKeys()

	if len(keys) != 3 {
		t.Fatalf("must be 3 keys not %v ", len(keys))
	}

	if keys[0] != "1" {
		t.Fatalf("1st key must be '1' not %v ", keys[0])
	}
	if keys[1] != "2" {
		t.Fatalf("2nd key must be '2' not %v ", keys[1])
	}
	if keys[2] != "3" {
		t.Fatalf("3rd key must be '3' not %v ", keys[2])
	}
}

func TestOrderedMap_DelHead(t *testing.T) {
	m := prepMap()

	m.Delete("1")
	keys := m.GetKeys()

	if len(keys) != 2 {
		t.Fatalf("must be 2 keys not %v ", len(keys))
	}

	if keys[0] != "2" {
		t.Fatalf("key must be '2' not %v ", keys[0])
	}
	if keys[1] != "3" {
		t.Fatalf("key must be '3' not %v ", keys[1])
	}
}

func TestOrderedMap_DelTail(t *testing.T) {
	m := prepMap()

	m.Delete("3")
	keys := m.GetKeys()

	if len(keys) != 2 {
		t.Fatalf("must be 2 keys not %v ", len(keys))
	}

	if keys[0] != "1" {
		t.Fatalf("1st key must be '1' not %v ", keys[0])
	}
	if keys[1] != "2" {
		t.Fatalf("2nd key must be '2' not %v ", keys[1])
	}
}

func TestOrderedMap_Del(t *testing.T) {
	m := prepMap()

	m.Delete("2")
	keys := m.GetKeys()

	if len(keys) != 2 {
		t.Fatalf("must be 2 keys not %v ", len(keys))
	}

	if keys[0] != "1" {
		t.Fatalf("1st key must be '1' not %v ", keys[0])
	}
	if keys[1] != "3" {
		t.Fatalf("2nd key must be '2' not %v ", keys[1])
	}
}
