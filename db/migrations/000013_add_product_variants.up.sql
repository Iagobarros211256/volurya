ALTER TABLE products
ADD COLUMN sales_type TEXT NOT NULL DEFAULT 'standard'
CHECK (sales_type IN ('standard', 'preorder'));

CREATE TABLE product_variants (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL
        REFERENCES products(id) ON DELETE CASCADE,
    label TEXT NOT NULL
        CHECK (length(btrim(label)) > 0),
    price_cents BIGINT NOT NULL
        CHECK (price_cents >= 0),
    display_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE (product_id, label)
);

CREATE INDEX idx_product_variants_order
ON product_variants (product_id, display_order, id);
