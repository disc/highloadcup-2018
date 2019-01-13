package main

import (
	"sync"

	"github.com/emirpasic/gods/maps/treemap"
)

type SafeTreemapIndex struct {
	v   map[string]*treemap.Map
	mux sync.Mutex
}

func NewSafeTreemapIndex() *SafeTreemapIndex {
	return &SafeTreemapIndex{v: map[string]*treemap.Map{}}
}

func (idx *SafeTreemapIndex) Exists(key string) bool {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	_, ok := idx.v[key]

	return ok
}

func (idx *SafeTreemapIndex) Update(key string, value *treemap.Map) {
	idx.mux.Lock()

	idx.v[key] = value

	idx.mux.Unlock()
}

func (idx *SafeTreemapIndex) Delete(key string) {
	idx.mux.Lock()

	delete(idx.v, key)

	idx.mux.Unlock()
}

func (idx *SafeTreemapIndex) Get(key string) *treemap.Map {
	idx.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer idx.mux.Unlock()

	return idx.v[key]
}
