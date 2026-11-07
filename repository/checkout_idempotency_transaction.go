package repository

import (
	"api/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

var ErrCheckoutIdempotencyConflict = errors.New(
	"checkout idempotency key reused with different request",
)

// CreateCheckoutOrderIdempotent cria um pedido com reserva de estoque,
// ou recupera o pedido existente sem descontar estoque novamente.
func (r *OrderRepository) CreateCheckoutOrderIdempotent(
	userID int,
	items []models.CheckoutItemInput,
	key string,
	fingerprint string,
) (*models.OrderDetail, bool, error) {
	if userID <= 0 || len(items) == 0 ||
		len(key) == 0 || len(key) > 128 ||
		len(fingerprint) != 64 {
		return nil, false, fmt.Errorf("invalid idempotent checkout parameters")
	}

	for _, item := range items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return nil, false, fmt.Errorf("invalid checkout item")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, fmt.Errorf("begin checkout: %w", err)
	}
	defer tx.Rollback()

	var orderID int

	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (
			user_id,
			total,
			status,
			checkout_idempotency_key,
			checkout_request_fingerprint
		)
		VALUES ($1, 0, 'pending', $2, $3)
		ON CONFLICT (user_id, checkout_idempotency_key)
		WHERE checkout_idempotency_key IS NOT NULL
		DO NOTHING
		RETURNING id
	`, userID, key, fingerprint).Scan(&orderID)

	reused := errors.Is(err, sql.ErrNoRows)

	if err != nil && !reused {
		return nil, false, fmt.Errorf("insert checkout order: %w", err)
	}

	if reused {
		var originalFingerprint string

		err = tx.QueryRowContext(ctx, `
			SELECT id, checkout_request_fingerprint
			FROM orders
			WHERE user_id = $1
			  AND checkout_idempotency_key = $2
			FOR UPDATE
		`, userID, key).Scan(
			&orderID,
			&originalFingerprint,
		)
		if err != nil {
			return nil, false, fmt.Errorf(
				"load existing checkout: %w", err,
			)
		}

		if originalFingerprint != fingerprint {
			return nil, false, ErrCheckoutIdempotencyConflict
		}
	} else {
		var totalCents int64

		for _, item := range items {
			var unitPrice float64

			err = tx.QueryRowContext(ctx, `
				SELECT price
				FROM products
				WHERE id = $1
			`, item.ProductID).Scan(&unitPrice)
			if err != nil {
				return nil, false, fmt.Errorf(
					"read product %d: %w", item.ProductID, err,
				)
			}

			result, err := tx.ExecContext(ctx, `
				UPDATE products
				SET stock = stock - $1
				WHERE id = $2 AND stock >= $1 AND sales_type = 'standard'
			`, item.Quantity, item.ProductID)
			if err != nil {
				return nil, false, fmt.Errorf(
					"reserve stock: %w", err,
				)
			}

			affected, err := result.RowsAffected()
			if err != nil {
				return nil, false, err
			}
			if affected != 1 {
				return nil, false, fmt.Errorf(
					"%w: product %d",
					ErrCheckoutInsufficientStock,
					item.ProductID,
				)
			}

			_, err = tx.ExecContext(ctx, `
				INSERT INTO order_items (
					order_id, product_id, quantity, unit_price
				)
				VALUES ($1, $2, $3, $4)
			`, orderID, item.ProductID, item.Quantity, unitPrice)
			if err != nil {
				return nil, false, fmt.Errorf(
					"insert checkout item: %w", err,
				)
			}

			totalCents += int64(math.Round(unitPrice*100)) *
				int64(item.Quantity)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE orders
			SET total = $1, updated_at = NOW()
			WHERE id = $2
		`, float64(totalCents)/100, orderID)
		if err != nil {
			return nil, false, fmt.Errorf(
				"update checkout total: %w", err,
			)
		}
	}

	detail := &models.OrderDetail{}
	var status, paymentStatus string
	var paymentIntent sql.NullString

	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, total, status,
		       payment_intent_id,
		       COALESCE(payment_status, 'pending'),
		       created_at, updated_at
		FROM orders WHERE id = $1
	`, orderID).Scan(
		&detail.ID,
		&detail.UserID,
		&detail.TotalPrice,
		&status,
		&paymentIntent,
		&paymentStatus,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	)
	if err != nil {
		return nil, false, fmt.Errorf(
			"load checkout order: %w", err,
		)
	}

	detail.Status = models.OrderStatus(status)
	detail.PaymentStatus = models.PaymentStatus(paymentStatus)

	if paymentIntent.Valid {
		detail.PaymentIntentID = paymentIntent.String
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, order_id, product_id,
		       quantity, unit_price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, orderID)
	if err != nil {
		return nil, false, fmt.Errorf(
			"load checkout items: %w", err,
		)
	}

	for rows.Next() {
		var item models.OrderItem

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.CreatedAt,
		); err != nil {
			rows.Close()
			return nil, false, err
		}

		detail.Items = append(detail.Items, item)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, false, err
	}
	if err := rows.Close(); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf(
			"commit idempotent checkout: %w", err,
		)
	}

	return detail, reused, nil
}
