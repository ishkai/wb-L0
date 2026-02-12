package consumer

import (
	"awesomeProject3/project/cache"
	"awesomeProject3/project/database"
	"awesomeProject3/project/model"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestHandleMessage_InvalidJSON_Commits_NoDB_NoCacheSet(t *testing.T) {
	db := &database.MockDB{}
	ca := &cache.MockCache{}

	cons := NewConsumer(db, ca, nil)

	msg := kafka.Message{Value: []byte("{not-valid-json")}

	commit := cons.HandleMessage(context.Background(), msg)

	if !commit {
		t.Fatalf("expected commit=true for invalid json")
	}
	if db.InsertCalls != 0 {
		t.Fatalf("expected db.InsertOrder not called, got %d", db.InsertCalls)
	}
	if ca.SetCalls != 0 {
		t.Fatalf("expected cache.Set not called, got %d", ca.SetCalls)
	}
}

func TestHandleMessage_ValidationError_Commits_NoDB_NoCacheSet(t *testing.T) {
	db := &database.MockDB{}
	ca := &cache.MockCache{}

	cons := NewConsumer(db, ca, nil)

	cons.validateFn = func(o *model.Order) error {
		return errors.New("bad order")
	}

	order := model.Order{OrderUID: "x"}
	data, _ := json.Marshal(order)

	msg := kafka.Message{Value: data}

	commit := cons.HandleMessage(context.Background(), msg)

	if !commit {
		t.Fatalf("expected commit=true for validation error (bad data)")
	}
	if db.InsertCalls != 0 {
		t.Fatalf("expected db.InsertOrder not called, got %d", db.InsertCalls)
	}
	if ca.SetCalls != 0 {
		t.Fatalf("expected cache.Set not called, got %d", ca.SetCalls)
	}
}

func TestHandleMessage_DBError_DoesNotCommit_NoCacheSet(t *testing.T) {
	db := &database.MockDB{}
	ca := &cache.MockCache{}

	cons := NewConsumer(db, ca, nil)

	cons.validateFn = func(o *model.Order) error { return nil }

	db.InsertOrderFunc = func(order model.Order) error {
		return errors.New("db down")
	}

	order := model.Order{OrderUID: "order-777"}
	data, _ := json.Marshal(order)

	msg := kafka.Message{Value: data}

	commit := cons.HandleMessage(context.Background(), msg)

	if commit {
		t.Fatalf("expected commit=false on db error (must retry)")
	}
	if db.InsertCalls != 1 {
		t.Fatalf("expected db.InsertOrder called once, got %d", db.InsertCalls)
	}
	if ca.SetCalls != 0 {
		t.Fatalf("expected cache.Set not called on db error, got %d", ca.SetCalls)
	}
}

func TestHandleMessage_Success_Commits_InsertsAndCaches(t *testing.T) {
	db := &database.MockDB{}
	ca := &cache.MockCache{}

	cons := NewConsumer(db, ca, nil)
	cons.validateFn = func(o *model.Order) error { return nil }

	db.InsertOrderFunc = func(order model.Order) error { return nil }

	order := model.Order{OrderUID: "ok-1"}
	data, _ := json.Marshal(order)

	msg := kafka.Message{Value: data}

	commit := cons.HandleMessage(context.Background(), msg)

	if !commit {
		t.Fatalf("expected commit=true on success")
	}
	if db.InsertCalls != 1 {
		t.Fatalf("expected db.InsertOrder called once, got %d", db.InsertCalls)
	}
	if ca.SetCalls != 1 {
		t.Fatalf("expected cache.Set called once, got %d", ca.SetCalls)
	}
	if ca.LastSetUID != "ok-1" {
		t.Fatalf("expected cache.Set uid=ok-1, got %q", ca.LastSetUID)
	}
}
