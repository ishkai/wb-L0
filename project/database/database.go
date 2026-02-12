package database

import (
	"awesomeProject3/project/model"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	InsertOrder(o model.Order) error
	GetOrder(id string) (model.Order, error)
	GetAllOrders() ([]model.Order, error)
	Close()
}

type Database struct {
	Pool *pgxpool.Pool
}

func NewDB(connection string) (DB, error) {
	pool, err := pgxpool.New(context.Background(), connection)
	if err != nil {
		return nil, err
	}
	db := &Database{Pool: pool}
	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return db, nil
}

func (db *Database) Close() { db.Pool.Close() }

func (db *Database) InsertOrder(o model.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	jsonData, err := json.Marshal(o)
	if err != nil {
		return err
	}
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cmdTag, err := tx.Exec(ctx,
		"INSERT INTO orders (order_uid, data) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		o.OrderUID, jsonData,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		log.Printf("Order with id=%s already exists, skipped insert", o.OrderUID)
	}

	return tx.Commit(ctx)
}

func (db *Database) GetOrder(id string) (model.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var jsonData []byte

	row := db.Pool.QueryRow(ctx, "SELECT data FROM orders WHERE order_uid = $1", id)
	if err := row.Scan(&jsonData); err != nil {
		return model.Order{}, err
	}

	var o model.Order

	if err := json.Unmarshal(jsonData, &o); err != nil {
		return model.Order{}, err
	}

	return o, nil
}

func (db *Database) GetAllOrders() ([]model.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.Pool.Query(ctx, "SELECT data FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order

	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			return nil, err
		}
		var o model.Order

		if err := json.Unmarshal(jsonData, &o); err != nil {
			return nil, err
		}
		orders = append(orders, o)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
