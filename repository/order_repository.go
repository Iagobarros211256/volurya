package repository

import (
	"api/models"
	"database/sql"
	"fmt"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(order models.Order) (int, error) {
	var id int
	err := r.db.QueryRow(
		`INSERT INTO orders (user_id, total, status)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		order.UserID, order.Total, order.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}
	return id, nil
}

func (r *OrderRepository) CreateOrderItems(orderID int, items []models.OrderItem) error {
	for _, item := range items {
		_, err := r.db.Exec(
			`INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			 VALUES ($1, $2, $3, $4)`,
			orderID, item.ProductID, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return fmt.Errorf("failed to create order item for product %d: %w", item.ProductID, err)
		}
	}
	return nil
}

func (r *OrderRepository) UpdateOrderPaymentIntent(id int, paymentIntentID, status string) error {
	_, err := r.db.Exec(
		`UPDATE orders
		 SET payment_intent_id = $1, payment_status = $2, updated_at = NOW()
		 WHERE id = $3`,
		paymentIntentID, status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update order payment intent: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdateOrderStatus(id int, status string) error {
	_, err := r.db.Exec(
		`UPDATE orders
		 SET status = $1, updated_at = NOW()
		 WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetByPaymentIntentID(paymentIntentID string) (*models.Order, error) {
	var order models.Order
	err := r.db.QueryRow(
		`SELECT id, user_id, total, status, payment_intent_id
		 FROM orders WHERE payment_intent_id = $1`,
		paymentIntentID,
	).Scan(&order.ID, &order.UserID, &order.Total, &order.Status, &order.ChargeID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}
