package main

import "sync"

type SafeIndex struct {
	v   map[interface{}]interface{}
	mux sync.Mutex
}

func NewSafeIndex() *SafeIndex {
	return &SafeIndex{v: map[interface{}]interface{}{}}
}

func (idx *SafeIndex) Exists(key interface{}) bool {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	_, ok := idx.v[key]

	return ok
}

func (idx *SafeIndex) Update(key interface{}, value interface{}) {
	idx.mux.Lock()

	idx.v[key] = value

	idx.mux.Unlock()
}

func (idx *SafeIndex) Delete(key interface{}) {
	idx.mux.Lock()

	delete(idx.v, key)

	idx.mux.Unlock()
}

func (idx *SafeIndex) Get(key interface{}) interface{} {
	idx.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer idx.mux.Unlock()

	return idx.v[key]
}
