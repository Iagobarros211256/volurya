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

/*

 Erros sem wrap em 3 das 4 funções
goreturn 0, err                    // CreateOrder
return err                       // UpdateOrderChargeID
return err                       // UpdateOrderStatus
Apenas UpdateOrderPaymentIntent não tem wrap, mas nenhuma retorna erro com contexto. Padronize:
goreturn 0, fmt.Errorf("failed to create order: %w", err)

🔴 CreateOrder insere product_id diretamente
Reflete o design problemático da migration e da model Order. Uma ordem com múltiplos produtos é impossível. Esse método deveria inserir em order_items dentro de uma transação.

🔴 UpdateOrderChargeID é método legado
gofunc (or *OrderRepository) UpdateOrderChargeID(id int, chargeID string) error {
    _, err := or.connection.Exec("UPDATE orders SET charge_id = $1 WHERE id = $2", chargeID, id)
charge_id é do PagSeguro — provedor anterior. O projeto migrou para Stripe com payment_intent_id. Esse método deveria ser removido ou o campo charge_id consolidado com payment_intent_id.

🔴 UpdateOrderStatus atualiza payment_status mas o parâmetro chama status
gofunc (or *OrderRepository) UpdateOrderStatus(id int, status string) error {
    _, err := or.connection.Exec(
        "UPDATE orders SET payment_status = $1...",
Nome enganoso — UpdateOrderStatus sugere que atualiza o status da ordem (pending, paid, cancelled), mas na verdade atualiza payment_status. Renomeie para UpdateOrderPaymentStatus ou atualize ambos os campos:
go"UPDATE orders SET status = $1, payment_status = $2, updated_at = NOW() WHERE id = $3"

🟡 Nenhum método para buscar ordens
Não há GetByID, GetByUserID ou GetByPaymentIntentID. O payment_usecase.go provavelmente precisa buscar ordens pelo payment_intent_id para processar webhooks — como está fazendo isso sem esse método?

🟡 status string em vez de models.OrderStatus
gofunc (or *OrderRepository) UpdateOrderStatus(id int, status string) error {
models.OrderStatus já existe como tipo — use-o para segurança em tempo de compilação:
gofunc (or *OrderRepository) UpdateOrderStatus(id int, status models.OrderStatus) error {

🟡 Sem interface
Padrão recorrente — sem interface para mock nos testes.

🟡 connection vs db — inconsistência de nomenclatura
CartRepository usa db, OrderRepository usa connection. Padronize para db em todo o projeto.


*/
