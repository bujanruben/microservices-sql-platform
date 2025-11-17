package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type CacheItem struct {
	Data      any
	Timestamp time.Time
	TTL       time.Duration
}

type Cache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Since(item.Timestamp) > item.TTL {
		delete(c.items, key)
		return nil, false
	}

	return item.Data, true
}

func (c *Cache) Set(key string, data interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = CacheItem{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

func GenerateCacheKey(city, date string, days int, unit, agg string) string {
	keyString := fmt.Sprintf("%s:%s:%d:%s:%s", city, date, days, unit, agg)
	hash := sha1.Sum([]byte(keyString))
	return hex.EncodeToString(hash[:])
}
