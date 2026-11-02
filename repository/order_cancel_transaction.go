package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CancelCheckoutOrderAndRestoreStock cancela um checkout sem
// pagamento iniciado e restaura o estoque atomicamente.
func (r *OrderRepository) CancelCheckoutOrderAndRestoreStock(
	orderID int,
) error {
	if orderID <= 0 {
		return fmt.Errorf("ID de pedido inválido")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	var status string
	var paymentIntentID sql.NullString

	err = tx.QueryRowContext(ctx, `
		SELECT status, payment_intent_id
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, orderID).Scan(&status, &paymentIntentID)

	if err != nil {
		return fmt.Errorf("bloquear pedido: %w", err)
	}

	// Cancelamento repetido não devolve estoque novamente.
	if status == "cancelled" {
		return nil
	}

	if status != "pending" {
		return fmt.Errorf(
			"pedido %d não pode ser cancelado: status %s",
			orderID, status,
		)
	}

	// Não cancelar pedidos vinculados à Stripe.
	if paymentIntentID.Valid {
		return fmt.Errorf(
			"pedido %d possui PaymentIntent vinculado",
			orderID,
		)
	}

	var hasPayment bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM payment_records
			WHERE order_id = $1
		)
	`, orderID).Scan(&hasPayment)
	if err != nil {
		return fmt.Errorf("verificar pagamentos: %w", err)
	}

	if hasPayment {
		return fmt.Errorf(
			"pedido %d possui registros financeiros",
			orderID,
		)
	}

	var productCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT product_id)
		FROM order_items
		WHERE order_id = $1
	`, orderID).Scan(&productCount)
	if err != nil {
		return fmt.Errorf("verificar itens: %w", err)
	}

	if productCount == 0 {
		return fmt.Errorf(
			"pedido %d não possui itens para restaurar",
			orderID,
		)
	}

	// Agrupa produtos repetidos e devolve suas quantidades.
	result, err := tx.ExecContext(ctx, `
		UPDATE products AS p
		SET stock = p.stock + items.quantity
		FROM (
			SELECT product_id, SUM(quantity) AS quantity
			FROM order_items
			WHERE order_id = $1
			GROUP BY product_id
		) AS items
		WHERE p.id = items.product_id
	`, orderID)
	if err != nil {
		return fmt.Errorf("restaurar estoque: %w", err)
	}

	restored, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("verificar restauração: %w", err)
	}

	if restored != int64(productCount) {
		return fmt.Errorf(
			"restauração incompleta: esperado=%d restaurado=%d",
			productCount, restored,
		)
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE orders
		SET status = 'cancelled',
		    updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, orderID)
	if err != nil {
		return fmt.Errorf("cancelar pedido: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("verificar cancelamento: %w", err)
	}

	if affected != 1 {
		return fmt.Errorf("pedido não foi cancelado")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar cancelamento: %w", err)
	}

	return nil
}
