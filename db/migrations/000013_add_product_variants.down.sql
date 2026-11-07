DROP TABLE IF EXISTS product_variants;

ALTER TABLE products
DROP COLUMN IF EXISTS sales_type;
