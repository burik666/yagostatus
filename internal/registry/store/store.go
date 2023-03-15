package store

import (
	"sync"
)

var m sync.Map

func Store(key, value interface{}) {
	m.Store(key, value)
}

func Load(key interface{}) (value interface{}, ok bool) {
	return m.Load(key)
}

func LoadAndDelete(key interface{}) (value interface{}, loaded bool) {
	return m.LoadAndDelete(key)
}

func LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	return m.LoadOrStore(key, value)
}

func Delete(key interface{}) {
	m.Delete(key)
}

func Range(f func(key, value interface{}) bool) {
	m.Range(f)
}
