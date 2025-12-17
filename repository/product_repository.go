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
// get all
func (pr *ProductRepository) GetProducts() ([]models.Product, error) {

	query := "SELECT id, name, description, price, stock FROM product"
	rows, err := pr.connection.Query(query)
	if err != nil {
		fmt.Println(err)
		return []models.Product{}, err
	}

	var productList []models.Product
	var productObj models.Product

	for rows.Next() {
		err = rows.Scan(
			&productObj.ID,
			&productObj.Name,
			&productObj.Description,
			&productObj.Price,
			&productObj.Stock)

		if err != nil {
			fmt.Println(err)
			return []models.Product{}, err
		}

		productList = append(productList, productObj)
	}

	rows.Close()

	return productList, nil
}

// create one        ?/in future whe will need a create many func/?
func (pr *ProductRepository) CreateProduct(product models.Product) (int, error) {

	var id int
	query, err := pr.connection.Prepare("INSERT INTO product" +
		"(name, description, price, stock)" +
		" VALUES ($1, $2, $3, $4) RETURNING id")
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	err = query.QueryRow(product.Name, product.Description, product.Price, product.Stock).Scan(&id)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	query.Close()
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
