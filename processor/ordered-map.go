package processor

type omItem struct {
	payload *QPayload
	key     string
	prev    *omItem
	next    *omItem
}

type omList struct {
	head *omItem
	tail *omItem
}

// Add adds element to the list
func (o *omList) Add(p *QPayload) *omItem {
	var i *omItem
	if o.head == nil {
		i = &omItem{
			payload: p,
		}
		o.head = i
		o.tail = o.head
	} else {
		i = &omItem{
			payload: p,
			prev:    o.tail,
		}
		o.tail.next = i
		o.tail = i
	}
	return i
}

// Del deletes element from the list
func (o *omList) Del(i *omItem) {

	if i.prev == nil && i.next == nil {
		o.head = nil
		o.tail = nil
		return
	}

	if i == o.head {
		o.head = i.next
	}
	if i == o.tail {
		o.tail = i.prev
	}
	if i.prev != nil {
		i.prev.next = i.next
	}
	if i.next != nil {
		i.next.prev = i.prev
	}

}

type OrderedMap struct {
	keys map[string]*omItem
	list omList
}

// NewOrderedMap creates a new ordered map
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		keys: make(map[string]*omItem),
	}
}

// Get finds element in the map
func (m *OrderedMap) Get(key string) (*QPayload, bool) {
	p, ok := m.keys[key]
	if ok {
		return p.payload, true
	}

	return nil, false
}

// FlattenWithKeys returns array of keys and array of elements
func (m *OrderedMap) FlattenWithKeys() ([]string, []*QPayload) {
	res := make([]*QPayload, len(m.keys))
	resKeys := make([]string, len(m.keys))
	j := 0
	for i := m.list.head; i != nil; i = i.next {
		resKeys[j] = i.key
		res[j] = i.payload
		j++
	}
	return resKeys, res

}

// Flatten returns array of elements
func (m *OrderedMap) Flatten() []*QPayload {
	res := make([]*QPayload, len(m.keys))
	j := 0
	for i := m.list.head; i != nil; i = i.next {
		res[j] = i.payload
		j++
	}
	return res
}

// GetKeys returns array of keys
func (m *OrderedMap) GetKeys() []string {
	res := make([]string, len(m.keys))
	j := 0
	for i := m.list.head; i != nil; i = i.next {
		res[j] = i.key
		j++
	}
	return res
}

// Add adds element to the map
func (m *OrderedMap) Add(key string, value *QPayload) {
	_, alreadyExist := m.keys[key]
	if alreadyExist {
		m.Delete(key)
	}

	i := m.list.Add(value)
	i.key = key
	m.keys[key] = i
}

// Delete deletes element from the map
func (m *OrderedMap) Delete(key string) bool {
	i, ok := m.keys[key]
	if ok {
		m.list.Del(i)
		delete(m.keys, key)
	}
	return true
}
