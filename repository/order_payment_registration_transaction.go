package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// RegisterPaymentIntentAtomically vincula um PaymentIntent ao pedido
// e cria seu registro financeiro na mesma transação PostgreSQL.
func (r *OrderRepository) RegisterPaymentIntentAtomically(
	orderID int,
	paymentIntentID string,
	amount int,
	currency string,
	userEmail string,
) error {
	if orderID <= 0 || paymentIntentID == "" ||
		amount <= 0 || currency != "BRL" {
		return fmt.Errorf("invalid payment registration parameters")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin payment registration: %w", err)
	}
	defer tx.Rollback()

	var status string
	var existingIntent sql.NullString
	var expectedAmount int64

	err = tx.QueryRowContext(ctx, `
		SELECT status,
		       payment_intent_id,
		       ROUND(total * 100)::BIGINT
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, orderID).Scan(
		&status,
		&existingIntent,
		&expectedAmount,
	)
	if err != nil {
		return fmt.Errorf("lock order for payment registration: %w", err)
	}

	if status != "pending" {
		return fmt.Errorf("order %d is not pending: %s", orderID, status)
	}

	if expectedAmount != int64(amount) {
		return fmt.Errorf(
			"payment amount mismatch: expected %d cents, got %d",
			expectedAmount, amount,
		)
	}

	if existingIntent.Valid {
		if existingIntent.String != paymentIntentID {
			return fmt.Errorf(
				"order %d already has another payment intent",
				orderID,
			)
		}

		// Permite repetir uma operação já concluída sem duplicar
		// o registro financeiro.
		var recordedAmount int
		var recordedCurrency, recordedEmail string

		err = tx.QueryRowContext(ctx, `
			SELECT amount, currency, user_email
			FROM payment_records
			WHERE order_id = $1
			  AND payment_intent_id = $2
		`, orderID, paymentIntentID).Scan(
			&recordedAmount,
			&recordedCurrency,
			&recordedEmail,
		)

		if err == nil {
			if recordedAmount != amount ||
				recordedCurrency != currency ||
				recordedEmail != userEmail {
				return fmt.Errorf(
					"existing payment record does not match request",
				)
			}
			return tx.Commit()
		}

		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check existing payment record: %w", err)
		}

		// O vínculo já existe, mas o registro financeiro não.
		// Continua para reparar esse estado incompleto.
	} else {
		result, err := tx.ExecContext(ctx, `
			UPDATE orders
			SET payment_intent_id = $1,
			    payment_status = 'pending',
			    updated_at = NOW()
			WHERE id = $2
			  AND status = 'pending'
			  AND payment_intent_id IS NULL
		`, paymentIntentID, orderID)
		if err != nil {
			return fmt.Errorf("link payment intent to order: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check order update: %w", err)
		}
		if affected != 1 {
			return fmt.Errorf("order payment link was not updated")
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment_records (
			order_id,
			payment_intent_id,
			amount,
			currency,
			status,
			user_email,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4,
			'requires_payment_method',
			$5, NOW(), NOW()
		)
	`, orderID, paymentIntentID, amount, currency, userEmail)
	if err != nil {
		return fmt.Errorf("insert payment record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit payment registration: %w", err)
	}

	return nil
}
