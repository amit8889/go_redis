package main

import "sync"

type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewKV() *KV {
	return &KV{
		data: map[string][]byte{},
	}
}

func (k *KV) Get(key string) []byte {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.data[key]
}

func (k *KV) Set(key string, val string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data[key] = []byte(val)
	return nil
}
