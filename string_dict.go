package main

import "sync"

func NewStringDictionary() *StringDictionary {
	return &StringDictionary{
		v: make(map[string]uint8, 0),
		k: make(map[uint8]string, 0),
	}
}

type StringDictionary struct {
	v map[string]uint8
	k map[uint8]string
	sync.Mutex
}

func (d *StringDictionary) Add(value string) uint8 {
	d.Lock()
	defer d.Unlock()

	if id, ok := d.v[value]; ok {
		return id
	}

	id := uint8(len(d.v)) + 1
	d.v[value] = id
	d.k[id] = value

	return id
}

func (d *StringDictionary) Get(id uint8) string {
	d.Lock()
	defer d.Unlock()

	if value, ok := d.k[id]; ok {
		return value
	}
	return ""
}

func (d *StringDictionary) GetId(value string) uint8 {
	d.Lock()
	defer d.Unlock()

	if id, ok := d.v[value]; ok {
		return id
	}
	return 0
}
