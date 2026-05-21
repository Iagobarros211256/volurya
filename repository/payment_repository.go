package repository

import (
	"api/models"
	"database/sql"
	"fmt"
	"time"
)

type PaymentRepository struct {
	connection *sql.DB
}

func NewPaymentRepository(connection *sql.DB) *PaymentRepository {
	return &PaymentRepository{connection: connection}
}

// CreatePaymentRecord creates a new payment record
func (pr *PaymentRepository) CreatePaymentRecord(payment models.PaymentRecord) (int, error) {
	var id int
	err := pr.connection.QueryRow(
		`INSERT INTO payment_records (
			order_id, payment_intent_id, amount, currency, status, stripe_customer_id, error_message, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		payment.OrderID,
		payment.PaymentIntentID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.StripeCustomerID,
		payment.ErrorMessage,
		time.Now(),
		time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetPaymentByIntentID retrieves a payment by its Stripe payment intent ID
func (pr *PaymentRepository) GetPaymentByIntentID(paymentIntentID string) (*models.PaymentRecord, error) {
	payment := &models.PaymentRecord{}
	err := pr.connection.QueryRow(
		`SELECT id, order_id, payment_intent_id, amount, currency, status, stripe_customer_id, error_message, created_at, updated_at
		 FROM payment_records WHERE payment_intent_id = $1`,
		paymentIntentID,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.PaymentIntentID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.StripeCustomerID,
		&payment.ErrorMessage,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment record not found")
		}
		return nil, err
	}
	return payment, nil
}

// GetPaymentByOrderID retrieves the latest payment for an order
func (pr *PaymentRepository) GetPaymentByOrderID(orderID int) (*models.PaymentRecord, error) {
	payment := &models.PaymentRecord{}
	err := pr.connection.QueryRow(
		`SELECT id, order_id, payment_intent_id, amount, currency, status, stripe_customer_id, error_message, created_at, updated_at
		 FROM payment_records WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1`,
		orderID,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.PaymentIntentID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.StripeCustomerID,
		&payment.ErrorMessage,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no payment records found for order")
		}
		return nil, err
	}
	return payment, nil
}

// UpdatePaymentStatus updates the status of a payment
func (pr *PaymentRepository) UpdatePaymentStatus(id int, status string) error {
	_, err := pr.connection.Exec(
		"UPDATE payment_records SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}

// UpdatePaymentStatusWithError updates the status and error message of a payment
func (pr *PaymentRepository) UpdatePaymentStatusWithError(id int, status, errorMessage string) error {
	_, err := pr.connection.Exec(
		"UPDATE payment_records SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3",
		status, errorMessage, id,
	)
	return err
}

// UpdateStripeCustomerID updates the Stripe customer ID for a payment
func (pr *PaymentRepository) UpdateStripeCustomerID(id int, customerID string) error {
	_, err := pr.connection.Exec(
		"UPDATE payment_records SET stripe_customer_id = $1, updated_at = NOW() WHERE id = $2",
		customerID, id,
	)
	return err
}

// ListPaymentsByStatus retrieves all payments with a specific status
func (pr *PaymentRepository) ListPaymentsByStatus(status string, limit, offset int) ([]models.PaymentRecord, error) {
	rows, err := pr.connection.Query(
		`SELECT id, order_id, payment_intent_id, amount, currency, status, stripe_customer_id, error_message, created_at, updated_at
		 FROM payment_records WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		status, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.PaymentRecord
	for rows.Next() {
		payment := models.PaymentRecord{}
		err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.PaymentIntentID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.StripeCustomerID,
			&payment.ErrorMessage,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, rows.Err()
}
