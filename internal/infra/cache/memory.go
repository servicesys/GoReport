package cache

import (
	"errors"
	"sync"
	"time"

	"reports-system/internal/domain/entities"
)

type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

func NewMemoryCache() entities.CacheProvider {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}

	// Cleanup goroutine
	go cache.cleanup()

	return cache
}

func (m *MemoryCache) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, errors.New("key not found")
	}

	if time.Now().After(item.expiresAt) {
		delete(m.data, key)
		return nil, errors.New("key expired")
	}

	return item.value, nil
}

func (m *MemoryCache) Set(key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (m *MemoryCache) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, item := range m.data {
				if now.After(item.expiresAt) {
					delete(m.data, key)
				}
			}
			m.mu.Unlock()
		}
	}
}
