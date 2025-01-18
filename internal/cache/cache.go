// internal/cache/cache.go

package cache

import (
	"sync"
	"time"

	"github.com/Lentz92/huggyfit/internal/calculator"
)

type CacheKey struct {
	ModelID    string
	Users      int
	ContextLen int
	DataType   calculator.DataType
}

type CacheEntry struct {
	Config    *calculator.ModelConfig
	KVCache   float64
	ExpiresAt time.Time
}

type Cache struct {
	configs      map[string]*calculator.ModelConfig
	calculations map[CacheKey]float64
	mu           sync.RWMutex
	expiration   time.Duration
}

func NewCache(expiration time.Duration) *Cache {
	return &Cache{
		configs:      make(map[string]*calculator.ModelConfig),
		calculations: make(map[CacheKey]float64),
		expiration:   expiration,
	}
}

func (c *Cache) GetConfig(modelID string) (*calculator.ModelConfig, bool) {
	c.mu.RLock()
	config, exists := c.configs[modelID]
	c.mu.RUnlock()
	return config, exists
}

func (c *Cache) SetConfig(modelID string, config *calculator.ModelConfig) {
	c.mu.Lock()
	c.configs[modelID] = config
	c.mu.Unlock()
}

func (c *Cache) GetKVCache(key CacheKey) (float64, bool) {
	c.mu.RLock()
	value, exists := c.calculations[key]
	c.mu.RUnlock()
	return value, exists
}

func (c *Cache) SetKVCache(key CacheKey, value float64) {
	c.mu.Lock()
	c.calculations[key] = value
	c.mu.Unlock()
}

// GetOrCalculateKVCache tries to get cached KV calculation or computes it if not found
func (c *Cache) GetOrCalculateKVCache(
	key CacheKey,
	parameters float64,
	useEstimation bool,
) float64 {
	// Try to get from cache first
	if cachedValue, exists := c.GetKVCache(key); exists {
		return cachedValue
	}

	var result float64
	if !useEstimation {
		// Try to get cached config
		config, exists := c.GetConfig(key.ModelID)
		if !exists {
			config, err := calculator.FetchModelConfig(key.ModelID)
			if err == nil {
				c.SetConfig(key.ModelID, config)

				kvParams := calculator.KVCacheParams{
					Users:         key.Users,
					ContextLength: key.ContextLen,
					DataType:      key.DataType,
					Config:        config,
				}

				result, err = calculator.CalculateKVCache(kvParams)
				if err == nil {
					c.SetKVCache(key, result)
					return result
				}
			}
		} else {
			kvParams := calculator.KVCacheParams{
				Users:         key.Users,
				ContextLength: key.ContextLen,
				DataType:      key.DataType,
				Config:        config,
			}

			var err error
			result, err = calculator.CalculateKVCache(kvParams)
			if err == nil {
				c.SetKVCache(key, result)
				return result
			}
		}
	}

	// Fallback to estimation
	result = calculator.EstimateKVCache(parameters, key.Users, key.ContextLen, key.DataType)
	c.SetKVCache(key, result)
	return result
}
