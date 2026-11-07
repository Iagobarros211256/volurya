DROP INDEX IF EXISTS orders_user_checkout_idempotency_uniq;

ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS orders_checkout_idempotency_pair_check,
    DROP COLUMN IF EXISTS checkout_idempotency_key,
    DROP COLUMN IF EXISTS checkout_request_fingerprint;
