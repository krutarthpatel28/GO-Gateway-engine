package cache

import (
	"fmt"
	"sync"
)

type MemoryCache struct{
	mu sync.RWMutex
	record map[string]map[string]string
}

func NewMemoryCache() *MemoryCache{
	return &MemoryCache{
		record: make(map[string]map[string]string),
	}
}

func (c *MemoryCache) Get(apiKey, key string) (string, error){
	c.mu.RLock()
	defer c.mu.RUnlock()
	userMap, exists := c.record[apiKey]
	if !exists{
		err := fmt.Errorf("User Not exists")
		return "", err
	}

	value, exists := userMap[key]
	if !exists{
		err := fmt.Errorf("Key does not exists")
		return "", err
	}
	return value, nil
}

func (c *MemoryCache) Set(apiKey, key, value string){
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.record[apiKey]; !exists{
		c.record[apiKey] = make(map[string]string)
	}
	c.record[apiKey][key] = value
}