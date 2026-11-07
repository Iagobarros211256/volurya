package repository

import (
	"api/models"
	"fmt"

	"github.com/lib/pq"
)

// GetVariantsByProductIDs busca variantes de vários produtos
// em uma única consulta, preservando a ordem de exibição.
func (pr *ProductRepository) GetVariantsByProductIDs(
	productIDs []int,
) (map[int][]models.ProductVariant, error) {
	result := make(map[int][]models.ProductVariant)

	if len(productIDs) == 0 {
		return result, nil
	}

	rows, err := pr.db.Query(`
		SELECT id, product_id, label, price_cents, display_order
		FROM product_variants
		WHERE product_id = ANY($1)
		ORDER BY product_id, display_order, id
	`, pq.Array(productIDs))
	if err != nil {
		return nil, fmt.Errorf("query product variants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var variant models.ProductVariant

		if err := rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.Label,
			&variant.PriceCents,
			&variant.DisplayOrder,
		); err != nil {
			return nil, fmt.Errorf("scan product variant: %w", err)
		}

		result[variant.ProductID] = append(
			result[variant.ProductID],
			variant,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product variants: %w", err)
	}

	return result, nil
}
