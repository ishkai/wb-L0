package database

import (
	"awesomeProject3/project/model"
	"sync"
)

type MockDB struct {
	mu sync.Mutex

	InsertOrderFunc  func(order model.Order) error
	GetOrderFunc     func(id string) (model.Order, error)
	GetAllOrdersFunc func() ([]model.Order, error)
	CloseFunc        func()

	InsertCalls int
	GetCalls    int
	GetAllCalls int
	CloseCalls  int

	LastInsert model.Order
	LastGetID  string
}

func (m *MockDB) InsertOrder(order model.Order) error {
	m.mu.Lock()
	m.InsertCalls++
	m.LastInsert = order
	m.mu.Unlock()

	if m.InsertOrderFunc != nil {
		return m.InsertOrderFunc(order)
	}
	return nil
}

func (m *MockDB) GetOrder(id string) (model.Order, error) {
	m.mu.Lock()
	m.GetCalls++
	m.LastGetID = id
	m.mu.Unlock()

	if m.GetOrderFunc != nil {
		return m.GetOrderFunc(id)
	}
	return model.Order{}, nil
}

func (m *MockDB) GetAllOrders() ([]model.Order, error) {
	m.mu.Lock()
	m.GetAllCalls++
	m.mu.Unlock()

	if m.GetAllOrdersFunc != nil {
		return m.GetAllOrdersFunc()
	}
	return nil, nil
}

func (m *MockDB) Close() {
	m.mu.Lock()
	m.CloseCalls++
	m.mu.Unlock()

	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
