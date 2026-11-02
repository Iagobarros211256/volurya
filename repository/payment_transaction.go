package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ConfirmPaymentSucceeded atualiza pagamento e pedido atomicamente.
// Retorna true quando houve alteração e false quando já estava concluído.
func (pr *PaymentRepository) ConfirmPaymentSucceeded(
	paymentIntentID string,
) (bool, error) {
	if paymentIntentID == "" {
		return false, fmt.Errorf("payment intent ID vazio")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := pr.connection.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	var paymentID, orderID int
	var paymentStatus, orderStatus string
	var orderPaymentStatus, linkedIntent sql.NullString

	err = tx.QueryRowContext(ctx, `
		SELECT p.id, p.order_id, p.status,
		       o.status, o.payment_status, o.payment_intent_id
		FROM payment_records p
		JOIN orders o ON o.id = p.order_id
		WHERE p.payment_intent_id = $1
		FOR UPDATE OF p, o
	`, paymentIntentID).Scan(
		&paymentID, &orderID, &paymentStatus,
		&orderStatus, &orderPaymentStatus, &linkedIntent,
	)
	if err != nil {
		return false, fmt.Errorf("localizar pagamento e pedido: %w", err)
	}

	if linkedIntent.Valid && linkedIntent.String != paymentIntentID {
		return false, fmt.Errorf("payment intent divergente no pedido")
	}

	switch orderStatus {
	case "pending", "failed", "paid":
	default:
		return false, fmt.Errorf("estado de pedido incompatível: %s", orderStatus)
	}

	switch paymentStatus {
	case "requires_payment_method", "requires_action",
		"processing", "failed", "succeeded":
	default:
		return false, fmt.Errorf("estado de pagamento incompatível: %s", paymentStatus)
	}

	alreadyComplete := paymentStatus == "succeeded" &&
		orderStatus == "paid" &&
		orderPaymentStatus.String == "succeeded" &&
		linkedIntent.Valid

	if alreadyComplete {
		return false, nil
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE payment_records
		SET status = 'succeeded',
		    error_message = '',
		    updated_at = NOW()
		WHERE id = $1
	`, paymentID)
	if err != nil {
		return false, fmt.Errorf("atualizar pagamento: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE orders
		SET status = 'paid',
		    payment_status = 'succeeded',
		    payment_intent_id = COALESCE(payment_intent_id, $2),
		    updated_at = NOW()
		WHERE id = $1
	`, orderID, paymentIntentID)
	if err != nil {
		return false, fmt.Errorf("atualizar pedido: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("confirmar transação: %w", err)
	}

	return true, nil
}

// RecordPaymentFailure registra uma tentativa recusada sem regredir
// pagamentos confirmados nem encerrar automaticamente o pedido.
func (pr *PaymentRepository) RecordPaymentFailure(
	paymentIntentID, errorMessage string,
) (bool, error) {
	if paymentIntentID == "" {
		return false, fmt.Errorf("payment intent ID vazio")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := pr.connection.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	var paymentID, orderID int
	var paymentStatus, orderStatus string
	var linkedIntent sql.NullString

	err = tx.QueryRowContext(ctx, `
		SELECT p.id, p.order_id, p.status,
		       o.status, o.payment_intent_id
		FROM payment_records p
		JOIN orders o ON o.id = p.order_id
		WHERE p.payment_intent_id = $1
		FOR UPDATE OF p, o
	`, paymentIntentID).Scan(
		&paymentID, &orderID, &paymentStatus,
		&orderStatus, &linkedIntent,
	)
	if err != nil {
		return false, fmt.Errorf("localizar pagamento e pedido: %w", err)
	}

	if linkedIntent.Valid && linkedIntent.String != paymentIntentID {
		return false, fmt.Errorf("payment intent divergente no pedido")
	}

	// Estados finais não podem ser revertidos por eventos atrasados.
	if paymentStatus == "succeeded" ||
		orderStatus == "paid" ||
		orderStatus == "cancelled" ||
		orderStatus == "refunded" {
		return false, nil
	}

	if orderStatus != "pending" {
		return false, fmt.Errorf(
			"pedido não elegível para falha: %s", orderStatus,
		)
	}

	if paymentStatus == "failed" {
		return false, nil
	}

	switch paymentStatus {
	case "requires_payment_method", "requires_action", "processing":
	default:
		return false, fmt.Errorf(
			"estado de pagamento incompatível: %s", paymentStatus,
		)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE payment_records
		SET status = 'failed',
		    error_message = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, paymentID, errorMessage)
	if err != nil {
		return false, fmt.Errorf("atualizar tentativa: %w", err)
	}

	// O pedido continua pendente para permitir nova tentativa.
	_, err = tx.ExecContext(ctx, `
		UPDATE orders
		SET payment_status = 'failed',
		    updated_at = NOW()
		WHERE id = $1
	`, orderID)
	if err != nil {
		return false, fmt.Errorf("atualizar estado financeiro: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("confirmar transação: %w", err)
	}

	return true, nil
}
