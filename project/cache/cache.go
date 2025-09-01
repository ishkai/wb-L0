package cache

import (
	"awesomeProject3/project/model"
	"sync"
	"time"
)

// Cache saving orders in memory.
type Cache struct {
	mu     sync.RWMutex
	orders map[string]cachedOrder
	ttl    time.Duration
}

type cachedOrder struct {
	order     model.Order
	timestamp time.Time
}

// New creating an empty cache.
func New(ttl time.Duration) *Cache {
	c := &Cache{
		orders: make(map[string]cachedOrder),
		ttl:    ttl,
	}

	go func() {
		ticker := time.NewTicker(ttl)
		defer ticker.Stop()

		for range ticker.C {
			c.mu.Lock()
			for id, co := range c.orders {
				if time.Since(co.timestamp) > c.ttl {
					delete(c.orders, id)
				}
			}
			c.mu.Unlock()
		}
	}()
	return c

}

// Get getting an order with id from cache.
func (c *Cache) Get(orderUID string) (model.Order, bool) {
	c.mu.RLock()
	co, ok := c.orders[orderUID]
	c.mu.RUnlock()

	if !ok {
		return model.Order{}, false
	}

	// Проверяем TTL
	if time.Since(co.timestamp) > c.ttl {
		c.mu.Lock()
		delete(c.orders, orderUID)
		c.mu.Unlock()
		return model.Order{}, false
	}

	return co.order, true
}

// Set setting order in cache with id.
func (c *Cache) Set(orderUID string, o model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[orderUID] = cachedOrder{
		order:     o,
		timestamp: time.Now(),
	}
}

// Delete deleting an order from cache.
func (c *Cache) Delete(orderUID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.orders, orderUID)
}
