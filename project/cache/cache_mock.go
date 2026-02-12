package cache

import (
	"awesomeProject3/project/model"
	"sync"
)

type MockCache struct {
	mu sync.Mutex

	GetFunc    func(orderUID string) (model.Order, bool)
	SetFunc    func(orderUID string, o model.Order)
	DeleteFunc func(orderUID string)

	GetCalls    int
	SetCalls    int
	DeleteCalls int

	LastGetUID    string
	LastSetUID    string
	LastSetOrder  model.Order
	LastDeleteUID string
}

func (m *MockCache) Get(orderUID string) (model.Order, bool) {
	m.mu.Lock()
	m.GetCalls++
	m.LastGetUID = orderUID
	m.mu.Unlock()

	if m.GetFunc != nil {
		return m.GetFunc(orderUID)
	}
	return model.Order{}, false
}

func (m *MockCache) Set(orderUID string, o model.Order) {
	m.mu.Lock()
	m.SetCalls++
	m.LastSetUID = orderUID
	m.LastSetOrder = o
	m.mu.Unlock()

	if m.SetFunc != nil {
		m.SetFunc(orderUID, o)
	}
}

func (m *MockCache) Delete(orderUID string) {
	m.mu.Lock()
	m.DeleteCalls++
	m.LastDeleteUID = orderUID
	m.mu.Unlock()

	if m.DeleteFunc != nil {
		m.DeleteFunc(orderUID)
	}
}
