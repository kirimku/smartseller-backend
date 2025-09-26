package tenant

import (
	"sync"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// TenantCache provides caching interface for tenant-related data
type TenantCache interface {
	// Storefront caching
	GetStorefront(key string) *entity.Storefront
	SetStorefront(key string, storefront *entity.Storefront, ttl time.Duration)
	InvalidateStorefront(key string)
	
	// Customer caching (optional for performance)
	GetCustomer(storefrontID, customerID string) *entity.Customer
	SetCustomer(storefrontID, customerID string, customer *entity.Customer, ttl time.Duration)
	InvalidateCustomer(storefrontID, customerID string)
	
	// Generic cache operations
	Clear()
	Size() int
}

// cacheItem represents a cached item with expiration
type cacheItem struct {
	data      interface{}
	expiresAt time.Time
}

// inMemoryCache is a simple in-memory cache implementation
type inMemoryCache struct {
	items        map[string]cacheItem
	mu           sync.RWMutex
	maxSize      int
	cleanupDone  chan bool
}

// NewInMemoryTenantCache creates a new in-memory tenant cache
func NewInMemoryTenantCache(maxSize int, cleanupInterval time.Duration) TenantCache {
	cache := &inMemoryCache{
		items:       make(map[string]cacheItem),
		maxSize:     maxSize,
		cleanupDone: make(chan bool, 1),
	}
	
	// Start cleanup goroutine
	go cache.startCleanup(cleanupInterval)
	
	return cache
}

// GetStorefront retrieves a storefront from cache
func (c *inMemoryCache) GetStorefront(key string) *entity.Storefront {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	cacheKey := "storefront:" + key
	item, exists := c.items[cacheKey]
	if !exists || time.Now().After(item.expiresAt) {
		return nil
	}
	
	if storefront, ok := item.data.(*entity.Storefront); ok {
		return storefront
	}
	
	return nil
}

// SetStorefront stores a storefront in cache
func (c *inMemoryCache) SetStorefront(key string, storefront *entity.Storefront, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if we need to make space
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}
	
	cacheKey := "storefront:" + key
	c.items[cacheKey] = cacheItem{
		data:      storefront,
		expiresAt: time.Now().Add(ttl),
	}
}

// InvalidateStorefront removes a storefront from cache
func (c *inMemoryCache) InvalidateStorefront(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	cacheKey := "storefront:" + key
	delete(c.items, cacheKey)
}

// GetCustomer retrieves a customer from cache
func (c *inMemoryCache) GetCustomer(storefrontID, customerID string) *entity.Customer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	cacheKey := "customer:" + storefrontID + ":" + customerID
	item, exists := c.items[cacheKey]
	if !exists || time.Now().After(item.expiresAt) {
		return nil
	}
	
	if customer, ok := item.data.(*entity.Customer); ok {
		return customer
	}
	
	return nil
}

// SetCustomer stores a customer in cache
func (c *inMemoryCache) SetCustomer(storefrontID, customerID string, customer *entity.Customer, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if we need to make space
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}
	
	cacheKey := "customer:" + storefrontID + ":" + customerID
	c.items[cacheKey] = cacheItem{
		data:      customer,
		expiresAt: time.Now().Add(ttl),
	}
}

// InvalidateCustomer removes a customer from cache
func (c *inMemoryCache) InvalidateCustomer(storefrontID, customerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	cacheKey := "customer:" + storefrontID + ":" + customerID
	delete(c.items, cacheKey)
}

// Clear removes all items from cache
func (c *inMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]cacheItem)
}

// Size returns the number of items in cache
func (c *inMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.items)
}

// evictOldest removes the oldest expired item, or the least recently used item
func (c *inMemoryCache) evictOldest() {
	if len(c.items) == 0 {
		return
	}
	
	now := time.Now()
	var oldestKey string
	var oldestTime time.Time
	
	// First, try to remove any expired items
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
			return
		}
		
		// Track oldest item as fallback
		if oldestKey == "" || item.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiresAt
		}
	}
	
	// If no expired items, remove the oldest one
	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// startCleanup runs periodic cleanup of expired items
func (c *inMemoryCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.cleanupDone:
			return
		}
	}
}

// cleanupExpired removes expired items from cache
func (c *inMemoryCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}

// Stop gracefully shuts down the cache
func (c *inMemoryCache) Stop() {
	select {
	case c.cleanupDone <- true:
	default:
	}
}

// RedisCache is a Redis-based cache implementation (placeholder)
type RedisCache struct {
	// Redis client and configuration would go here
	// This is a placeholder for future Redis integration
}

// NewRedisTenantCache creates a new Redis-based tenant cache (placeholder)
func NewRedisTenantCache(redisURL string) TenantCache {
	// This would initialize a Redis client and return a Redis-based cache
	// For now, return in-memory cache as fallback
	return NewInMemoryTenantCache(1000, 5*time.Minute)
}