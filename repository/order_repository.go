package repository

import (
	"api/models"
	"database/sql"
)

type OrderRepository struct {
	connection *sql.DB
}

func NewOrderRepository(connection *sql.DB) *OrderRepository {
	return &OrderRepository{connection: connection}
}

func (or *OrderRepository) CreateOrder(order models.Order) (int, error) {
	var id int
	err := or.connection.QueryRow(
		"INSERT INTO orders (user_id, product_id, quantity, total, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		order.UserID, order.ProductID, order.Quantity, order.Total, order.Status,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (or *OrderRepository) UpdateOrderChargeID(id int, chargeID string) error {
	_, err := or.connection.Exec("UPDATE orders SET charge_id = $1 WHERE id = $2", chargeID, id)
	return err
}

// UpdateOrderPaymentIntent updates order with payment intent ID and status
func (or *OrderRepository) UpdateOrderPaymentIntent(id int, paymentIntentID, status string) error {
	_, err := or.connection.Exec(
		"UPDATE orders SET payment_intent_id = $1, payment_status = $2, updated_at = NOW() WHERE id = $3",
		paymentIntentID, status, id,
	)
	return err
}

// UpdateOrderStatus updates the payment status of an order
func (or *OrderRepository) UpdateOrderStatus(id int, status string) error {
	_, err := or.connection.Exec(
		"UPDATE orders SET payment_status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}
