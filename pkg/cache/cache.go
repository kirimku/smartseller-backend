// Package cache provides a generic caching mechanism for the application
package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(key string) (interface{}, bool)

	// Set stores a value in the cache with an optional expiration
	Set(key string, value interface{}, expiration ...time.Duration)

	// Delete removes a value from the cache
	Delete(key string)

	// Flush removes all items from the cache
	Flush()
}

// InMemoryCache implements Cache using go-cache
type InMemoryCache struct {
	cache *cache.Cache
}

// NewInMemoryCache creates a new in-memory cache with the specified default expiration
// and cleanup interval
func NewInMemoryCache(defaultExpiration, cleanupInterval time.Duration) *InMemoryCache {
	return &InMemoryCache{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(key string, value interface{}, expiration ...time.Duration) {
	// If expiration is provided, use it; otherwise, use default expiration
	exp := cache.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	c.cache.Set(key, value, exp)
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.cache.Delete(key)
}

// Flush removes all items from the cache
func (c *InMemoryCache) Flush() {
	c.cache.Flush()
}

// NoExpiration represents no expiration time
var NoExpiration = cache.NoExpiration

// DefaultExpiration represents the default expiration time
var DefaultExpiration = cache.DefaultExpiration
