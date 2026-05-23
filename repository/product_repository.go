package repository

import (
	"api/models"
	"database/sql"
	"errors"
	"fmt"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (pr *ProductRepository) GetProducts(limit int, cursor *int) ([]models.Product, bool, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var query string
	var args []any

	if cursor != nil {
		query = `SELECT id, user_id, name, description, price, stock, COALESCE(image_url, '')
				 FROM products WHERE id > $1 ORDER BY id ASC LIMIT $2`
		args = []any{*cursor, limit + 1}
	} else {
		query = `SELECT id, user_id, name, description, price, stock, COALESCE(image_url, '')
				 FROM products ORDER BY id ASC LIMIT $1`
		args = []any{limit + 1}
	}

	rows, err := pr.db.Query(query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query products failed: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0, limit)

	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
			&p.ImageURL,
		); err != nil {
			return nil, false, fmt.Errorf("scan failed: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("rows error: %w", err)
	}

	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}

	return products, hasMore, nil
}

func (pr *ProductRepository) CreateProduct(product models.Product, userID int) (int, error) {
	var id int
	err := pr.db.QueryRow(
		`INSERT INTO products (user_id, name, description, price, stock)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, product.Name, product.Description, product.Price, product.Stock,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create product failed: %w", err)
	}
	return id, nil
}

func (pr *ProductRepository) GetProductById(id int) (*models.Product, error) {
	var p models.Product
	err := pr.db.QueryRow(
		`SELECT id, user_id, name, description, price, stock, COALESCE(image_url, '')
		 FROM products WHERE id = $1`,
		id,
	).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description,
		&p.Price, &p.Stock, &p.ImageURL,
	)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product %d failed: %w", id, err)
	}
	return &p, nil
}

func (pr *ProductRepository) UpdateProduct(id int, name, description string, price float64, stock int) (*models.Product, error) {
	var p models.Product
	err := pr.db.QueryRow(
		`UPDATE products SET name = $1, description = $2, price = $3, stock = $4
		 WHERE id = $5
		 RETURNING id, user_id, name, description, price, stock, COALESCE(image_url, '')`,
		name, description, price, stock, id,
	).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description,
		&p.Price, &p.Stock, &p.ImageURL,
	)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update product %d failed: %w", id, err)
	}
	return &p, nil
}

func (pr *ProductRepository) Delete(id int) error {
	var deletedID int
	err := pr.db.QueryRow(
		`DELETE FROM products WHERE id = $1 RETURNING id`,
		id,
	).Scan(&deletedID)
	if err == sql.ErrNoRows {
		return ErrProductNotFound
	}
	if err != nil {
		return fmt.Errorf("delete product %d failed: %w", id, err)
	}
	return nil
}

func (pr *ProductRepository) UpdateImageURL(productID int, imageURL string) (*models.Product, error) {
	var p models.Product
	err := pr.db.QueryRow(
		`UPDATE products SET image_url = $1 WHERE id = $2
		 RETURNING id, user_id, name, description, price, stock, COALESCE(image_url, '')`,
		imageURL, productID,
	).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description,
		&p.Price, &p.Stock, &p.ImageURL,
	)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update image url failed: %w", err)
	}
	return &p, nil
}

// DecrementStock decrementa o estoque atomicamente.
// Retorna ErrProductNotFound se o produto não existir ou estoque insuficiente.
func (pr *ProductRepository) DecrementStock(productID, quantity int) error {
	result, err := pr.db.Exec(
		`UPDATE products
		 SET stock = stock - $1
		 WHERE id = $2 AND stock >= $1`,
		quantity, productID,
	)
	if err != nil {
		return fmt.Errorf("decrement stock failed for product %d: %w", productID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: product %d has insufficient stock", ErrProductNotFound, productID)
	}

	return nil
}
