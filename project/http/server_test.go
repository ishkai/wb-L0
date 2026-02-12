package http

import (
	"awesomeProject3/project/cache"
	"awesomeProject3/project/database"
	"awesomeProject3/project/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetOrderByPath_CacheHit_DBNotCalled(t *testing.T) {
	db := &database.MockDB{}
	c := &cache.MockCache{}

	want := model.Order{OrderUID: "order-1"}
	c.GetFunc = func(uid string) (model.Order, bool) {
		if uid != "order-1" {
			t.Fatalf("expected uid=order-1, got %q", uid)
		}
		return want, true
	}

	s := NewServer(db, c)

	req := httptest.NewRequest(http.MethodGet, "/order/order-1", nil)
	req = mux.SetURLVars(req, map[string]string{"order_uid": "order-1"})
	rr := httptest.NewRecorder()

	s.GetOrderByPath(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	if db.GetCalls != 0 {
		t.Fatalf("expected db.GetOrder not called, got %d calls", db.GetCalls)
	}

	if c.GetCalls != 1 {
		t.Fatalf("expected cache.Get called once, got %d", c.GetCalls)
	}

	if c.SetCalls != 0 {
		t.Fatalf("expected cache.Set not called on hit, got %d", c.SetCalls)
	}

	var got model.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if got.OrderUID != want.OrderUID {
		t.Fatalf("expected OrderUID %q, got %q", want.OrderUID, got.OrderUID)
	}
}

func TestGetOrderByPath_CacheMiss_DBCalled_ThenCacheSet(t *testing.T) {
	db := &database.MockDB{}
	c := &cache.MockCache{}

	c.GetFunc = func(uid string) (model.Order, bool) {
		return model.Order{}, false
	}

	want := model.Order{OrderUID: "order-2"}
	db.GetOrderFunc = func(id string) (model.Order, error) {
		if id != "order-2" {
			t.Fatalf("expected id=order-2, got %q", id)
		}
		return want, nil
	}

	s := NewServer(db, c)

	req := httptest.NewRequest(http.MethodGet, "/order/order-2", nil)
	req = mux.SetURLVars(req, map[string]string{"order_uid": "order-2"})
	rr := httptest.NewRecorder()

	s.GetOrderByPath(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	if c.GetCalls != 1 {
		t.Fatalf("expected cache.Get called once, got %d", c.GetCalls)
	}

	if db.GetCalls != 1 {
		t.Fatalf("expected db.GetOrder called once, got %d", db.GetCalls)
	}

	if c.SetCalls != 1 {
		t.Fatalf("expected cache.Set called once, got %d", c.SetCalls)
	}
	if c.LastSetUID != "order-2" {
		t.Fatalf("expected cache.Set uid=order-2, got %q", c.LastSetUID)
	}

	var got model.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if got.OrderUID != want.OrderUID {
		t.Fatalf("expected OrderUID %q, got %q", want.OrderUID, got.OrderUID)
	}
}
