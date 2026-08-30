// Package repository is the only part of order-service allowed to talk to
// the orders PostgreSQL database directly - no other service, and no other
// package within order-service, issues SQL against these tables.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shopstream/order-service/internal/models"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error
}

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if committed

	now := time.Now().UTC()
	order.CreatedAt = now
	order.UpdatedAt = now

	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, user_id, status, total_cents, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.ID, order.UserID, order.Status, order.TotalCents, order.Currency, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for i := range order.Items {
		item := &order.Items[i]
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (id, order_id, product_id, sku, quantity, unit_price_cents)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			item.ID, order.ID, item.ProductID, item.SKU, item.Quantity, item.UnitPriceCents,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	order := &models.Order{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total_cents, currency, created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalCents, &order.Currency, &order.CreatedAt, &order.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	items, err := r.getItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (r *PostgresOrderRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, status, total_cents, currency, created_at, updated_at
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalCents, &o.Currency, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now().UTC(), id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *PostgresOrderRepository) getItems(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, sku, quantity, unit_price_cents FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.SKU, &item.Quantity, &item.UnitPriceCents); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
