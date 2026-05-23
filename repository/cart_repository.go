package repository

import (
	"api/models"
	"database/sql"
	"fmt"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

// GetOrCreateCart busca o carrinho do usuário ou cria um novo
func (r *CartRepository) GetOrCreateCart(userID int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.QueryRow(
		"SELECT id, user_id, created_at FROM carts WHERE user_id = $1",
		userID,
	).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt)

	if err == sql.ErrNoRows {
		err = r.db.QueryRow(
			"INSERT INTO carts (user_id) VALUES ($1) RETURNING id, user_id, created_at",
			userID,
		).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create cart: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	return &cart, nil
}

// GetCartItems busca os itens do carrinho com detalhes do produto
func (r *CartRepository) GetCartItems(cartID int) ([]models.CartItem, error) {
	rows, err := r.db.Query(`
		SELECT 
			ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.added_at,
			p.id, p.user_id, p.name, p.description, p.price, p.stock
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.added_at ASC
	`, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	defer rows.Close()

	items := make([]models.CartItem, 0)
	for rows.Next() {
		var item models.CartItem
		var product models.Product
		err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.AddedAt,
			&product.ID, &product.UserID, &product.Name, &product.Description, &product.Price, &product.Stock,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		item.Product = &product
		items = append(items, item)
	}

	return items, nil
}

// AddItem adiciona um item ao carrinho ou atualiza a quantidade se já existir
func (r *CartRepository) AddItem(cartID, productID, quantity int) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.QueryRow(`
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
		RETURNING id, cart_id, product_id, quantity, added_at
	`, cartID, productID, quantity).Scan(
		&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.AddedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}
	return &item, nil
}

// UpdateItemQuantity atualiza a quantidade de um item
func (r *CartRepository) UpdateItemQuantity(itemID, quantity int) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.QueryRow(`
		UPDATE cart_items SET quantity = $1
		WHERE id = $2
		RETURNING id, cart_id, product_id, quantity, added_at
	`, quantity, itemID).Scan(
		&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.AddedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}
	return &item, nil
}

// RemoveItem remove um item do carrinho
func (r *CartRepository) RemoveItem(itemID int) error {
	_, err := r.db.Exec("DELETE FROM cart_items WHERE id = $1", itemID)
	if err != nil {
		return fmt.Errorf("failed to remove item: %w", err)
	}
	return nil
}

// ClearCart remove todos os itens do carrinho
func (r *CartRepository) ClearCart(cartID int) error {
	_, err := r.db.Exec("DELETE FROM cart_items WHERE cart_id = $1", cartID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}
	return nil
}

// GetItemByID busca um item pelo ID
func (r *CartRepository) GetItemByID(itemID int) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.QueryRow(
		"SELECT id, cart_id, product_id, quantity, added_at FROM cart_items WHERE id = $1",
		itemID,
	).Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.AddedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	return &item, nil
}

/*


GetOrCreateCart tem race condition
go// SELECT — não existe
// INSERT — outro request pode inserir entre o SELECT e o INSERT
Dois requests simultâneos do mesmo usuário podem passar pelo SELECT (ambos com ErrNoRows) e tentar o INSERT — o segundo vai falhar com violação de UNIQUE. Use upsert atômico:
goerr := r.db.QueryRow(`
    INSERT INTO carts (user_id) VALUES ($1)
    ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
    RETURNING id, user_id, created_at
`, userID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt)

🔴 UpdateItemQuantity não valida ownership
goUPDATE cart_items SET quantity = $1 WHERE id = $2
Qualquer usuário autenticado pode atualizar o item de outro usuário se souber o itemID. Adicione validação via cartID:
goUPDATE cart_items SET quantity = $1
WHERE id = $2 AND cart_id = $3  -- cart_id pertence ao usuário

🔴 RemoveItem também sem validação de ownership
goDELETE FROM cart_items WHERE id = $1
Mesmo problema — sem verificar se o item pertence ao carrinho do usuário:
goDELETE FROM cart_items WHERE id = $1 AND cart_id = $2

🟡 GetCartItems não retorna image_url do produto
goSELECT p.id, p.user_id, p.name, p.description, p.price, p.stock
image_url não está no SELECT mas está no model Product. O frontend não consegue exibir imagens dos itens do carrinho.

🟡 GetOrCreateCart não carrega os itens
goreturn &cart, nil  // cart.Items é nil
O caller precisa chamar GetCartItems separadamente. Isso não é necessariamente errado, mas o Cart retornado tem Items: nil em vez de Items: []CartItem{} — pode causar null em vez de [] no JSON.

🟡 Sem interface definida
Padrão recorrente do projeto — CartRepository é struct concreta sem interface, impedindo mocks nos testes.

🟢 items := make([]CartItem, 0) correto
Inicializar com make em vez de var items []CartItem garante que serializa como [] em vez de null no JSON quando vazio. Boa prática.


*/
