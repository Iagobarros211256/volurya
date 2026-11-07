package repository

import (
	"api/models"
	"context"
	"errors"
	"fmt"
)

var ErrCheckoutInsufficientStock = errors.New("checkout insufficient stock")

// CreateCheckoutOrder persiste o pedido, os itens e os descontos
// de estoque em uma única transação PostgreSQL.
func (r *OrderRepository) CreateCheckoutOrder(
	order models.Order,
	items []models.OrderItem,
) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("checkout sem itens")
	}

	for _, item := range items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return 0, fmt.Errorf(
				"item inválido: produto %d, quantidade %d",
				item.ProductID, item.Quantity,
			)
		}
	}

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	var orderID int

	err = tx.QueryRow(
		`INSERT INTO orders (user_id, total, status)
		 VALUES ($1, $2, $3) RETURNING id`,
		order.UserID, order.Total, order.Status,
	).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("criar pedido: %w", err)
	}

	for _, item := range items {
		_, err := tx.Exec(
			`INSERT INTO order_items
			 (order_id, product_id, quantity, unit_price)
			 VALUES ($1, $2, $3, $4)`,
			orderID, item.ProductID, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return 0, fmt.Errorf("inserir item: %w", err)
		}

		result, err := tx.Exec(
			`UPDATE products
			 SET stock = stock - $1
			 WHERE id = $2 AND stock >= $1 AND sales_type = 'standard'`,
			item.Quantity, item.ProductID,
		)
		if err != nil {
			return 0, fmt.Errorf("descontar estoque: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("verificar desconto: %w", err)
		}
		if rows != 1 {
			return 0, fmt.Errorf(
				"%w: estoque insuficiente para produto %d",
				ErrCheckoutInsufficientStock, item.ProductID,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("confirmar checkout: %w", err)
	}

	return orderID, nil
}
