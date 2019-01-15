package main

import (
	"sync"

	"github.com/emirpasic/gods/maps/treemap"
)

type SafeTreemapIndex struct {
	v   map[interface{}]*treemap.Map
	mux sync.Mutex
}

func NewSafeTreemapIndex() *SafeTreemapIndex {
	return &SafeTreemapIndex{v: map[interface{}]*treemap.Map{}}
}

func (idx *SafeTreemapIndex) Exists(key interface{}) bool {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	_, ok := idx.v[key]

	return ok
}

func (idx *SafeTreemapIndex) Update(key interface{}, value *treemap.Map) {
	idx.mux.Lock()

	idx.v[key] = value

	idx.mux.Unlock()
}

func (idx *SafeTreemapIndex) Delete(key interface{}) {
	idx.mux.Lock()

	delete(idx.v, key)

	idx.mux.Unlock()
}

func (idx *SafeTreemapIndex) Get(key interface{}) *treemap.Map {
	idx.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer idx.mux.Unlock()

	return idx.v[key]
}
