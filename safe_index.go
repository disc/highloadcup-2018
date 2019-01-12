package main

import "sync"

type SafeIndex struct {
	v   map[interface{}]interface{}
	mux sync.Mutex
}

func (idx *SafeIndex) Exists(key string) bool {
	idx.mux.Lock()
	defer idx.mux.Unlock()

	_, ok := idx.v[key]

	return ok
}

func (idx *SafeIndex) Update(key string, value interface{}) {
	idx.mux.Lock()

	idx.v[key] = value

	idx.mux.Unlock()
}

func (idx *SafeIndex) Delete(key string) {
	idx.mux.Lock()

	delete(idx.v, key)

	idx.mux.Unlock()
}
