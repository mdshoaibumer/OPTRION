-- 000002_create_products.down.sql
DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
DROP TABLE IF EXISTS products;
