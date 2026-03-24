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
