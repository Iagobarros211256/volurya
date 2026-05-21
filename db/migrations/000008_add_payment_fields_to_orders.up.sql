-- Add Stripe payment fields to orders table
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS payment_intent_id VARCHAR(255) UNIQUE,
ADD COLUMN IF NOT EXISTS payment_status VARCHAR(50) DEFAULT 'pending',
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Create payment_records table to track Stripe transactions
CREATE TABLE IF NOT EXISTS payment_records (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payment_intent_id VARCHAR(255) UNIQUE NOT NULL,
    amount INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'BRL',
    status VARCHAR(50) NOT NULL DEFAULT 'requires_payment_method',
    stripe_customer_id VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_payment_records_order_id ON payment_records(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_payment_intent_id ON payment_records(payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_orders_payment_intent_id ON orders(payment_intent_id);


/*


payment_status sem constraint de valores válidos
sqlpayment_status VARCHAR(50) DEFAULT 'pending'
Igual ao status em orders — qualquer string é aceita. Os status do Stripe são conhecidos:
sqlpayment_status VARCHAR(50) DEFAULT 'pending' 
    CHECK (payment_status IN (
        'pending', 'requires_payment_method', 'requires_confirmation',
        'requires_action', 'processing', 'succeeded', 'cancelled', 'failed'
    ))

🔴 status em payment_records também sem constraint
sqlstatus VARCHAR(50) NOT NULL DEFAULT 'requires_payment_method'
Mesma correção necessária.

🔴 amount INTEGER para valor monetário
sqlamount INTEGER NOT NULL
orders.total usa NUMERIC(10,2) corretamente, mas payment_records.amount usa INTEGER. Inconsistência perigosa — se alguém comparar os dois valores diretamente, a comparação falha. Escolha um padrão:

NUMERIC(10,2) em reais — consistente com orders.total
INTEGER em centavos — consistente com a API do Stripe

Mas documente explicitamente qual é e seja consistente:
sqlamount INTEGER NOT NULL CHECK (amount > 0), -- valor em centavos (Stripe)

🟡 updated_at adicionado em orders mas sem trigger
sqlADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
Mesmo problema da migration de products — o valor nunca é atualizado automaticamente em UPDATEs. O trigger sugerido anteriormente deveria ser criado aqui também.

🟡 payment_intent_id UNIQUE em ambas as tabelas
orders.payment_intent_id e payment_records.payment_intent_id são ambos UNIQUE. Isso impede reembolsos parciais ou múltiplas tentativas de pagamento para o mesmo pedido, que são casos reais no Stripe. Considere remover o UNIQUE de payment_records.payment_intent_id ou criar um índice composto:
sqlUNIQUE (order_id, payment_intent_id)

🟡 ON DELETE CASCADE em payment_records
sqlorder_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE
Deletar uma ordem apaga o histórico financeiro — problemático para auditoria e reconciliação contábil. Use ON DELETE RESTRICT:
sqlorder_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE RESTRICT

🟢 stripe_customer_id nullable sem índice
Se houver busca por customer futuramente:
sqlCREATE INDEX idx_payment_records_stripe_customer_id 
    ON payment_records(stripe_customer_id) 
    WHERE stripe_customer_id IS NOT NULL;


*/