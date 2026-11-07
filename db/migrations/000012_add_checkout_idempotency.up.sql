-- Identifica tentativas de checkout de um usuario.
-- Pedidos antigos permanecem compativeis com NULL.

ALTER TABLE orders
    ADD COLUMN checkout_idempotency_key VARCHAR(128),
    ADD COLUMN checkout_request_fingerprint VARCHAR(64);

-- Uma chave deve estar acompanhada do hash dos dados
-- originais da requisicao.

ALTER TABLE orders
    ADD CONSTRAINT orders_checkout_idempotency_pair_check
    CHECK (
        (
            checkout_idempotency_key IS NULL
            AND checkout_request_fingerprint IS NULL
        )
        OR
        (
            checkout_idempotency_key IS NOT NULL
            AND LENGTH(TRIM(checkout_idempotency_key)) > 0
            AND checkout_request_fingerprint ~ '^[0-9a-f]{64}$'
        )
    );

-- Cada usuario pode utilizar uma chave apenas uma vez.
-- Pedidos antigos, sem chave, nao sao afetados.

CREATE UNIQUE INDEX orders_user_checkout_idempotency_uniq
    ON orders (user_id, checkout_idempotency_key)
    WHERE checkout_idempotency_key IS NOT NULL;
