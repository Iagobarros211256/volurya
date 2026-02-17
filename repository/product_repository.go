package repository

import (
	"api/models"
	"database/sql"
	"fmt"
)

// classe principal construtora
type ProductRepository struct {
	connection *sql.DB
}

// funcao construtora
func NewProductRepository(connection *sql.DB) ProductRepository {
	return ProductRepository{
		connection: connection,
	}
}

// crud
// get all com pagination by a cursor (or indexer)
func (pr *ProductRepository) GetProducts(
	limit int,
	cursor *int,
) ([]models.Product, bool, error) {

	query := `
		SELECT id, name, description, price, stock
		FROM product
	`

	args := []any{}
	paramIndex := 1

	if cursor != nil {
		query += fmt.Sprintf(" WHERE id > $%d", paramIndex)
		args = append(args, *cursor)
		paramIndex++
	}

	query += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", paramIndex)
	args = append(args, limit+1)

	rows, err := pr.connection.Query(query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	products := []models.Product{}

	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
		); err != nil {
			return nil, false, err
		}
		products = append(products, p)
	}

	hasMore := false
	if len(products) > limit {
		hasMore = true
		products = products[:limit]
	}

	return products, hasMore, nil
}

func (pr *ProductRepository) CreateProduct(product models.Product, userID int) (int, error) {
	var id int
	err := pr.connection.QueryRow(
		"INSERT INTO products (user_id, name, description, price, stock) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID, product.Name, product.Description, product.Price, product.Stock,
	).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

// get one by id
func (pr *ProductRepository) GetProductById(id_product int) (*models.Product, error) {

	query, err := pr.connection.Prepare("SELECT * FROM product WHERE id = $1")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var produto models.Product

	err = query.QueryRow(id_product).Scan(
		&produto.ID,
		&produto.Name,
		&produto.Description,
		&produto.Price,
		&produto.Stock,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	query.Close()
	return &produto, nil
}

// update one        /? in future we will need a update many?/
func (pr *ProductRepository) UpdateProduct(id_product int, Name string, Description string, Price float64, Stock int) (*models.Product, error) {

	query, err := pr.connection.Prepare("UPDATE product SET name = $1, description = $2, price = $3, stock = $4 WHERE id = $5 RETURNING id, name, description, price, stock;")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var produto models.Product

	err = query.QueryRow(Name, Description, Price, Stock, id_product).Scan(
		&produto.ID,
		&produto.Name,
		&produto.Description,
		&produto.Price,
		&produto.Stock,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	query.Close()
	return &produto, nil
}

// delete one
func (pr *ProductRepository) Delete(id int) error {
	query := `DELETE FROM product WHERE id = $1 RETURNING id`

	var deletedID int
	err := pr.connection.
		QueryRow(query, id).
		Scan(&deletedID)

	if err == sql.ErrNoRows {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}

	return nil
}
