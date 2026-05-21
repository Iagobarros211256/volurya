-- Rollback payment migration
DROP INDEX IF EXISTS idx_orders_payment_intent_id;
DROP INDEX IF EXISTS idx_payment_records_payment_intent_id;
DROP INDEX IF EXISTS idx_payment_records_order_id;
DROP TABLE IF EXISTS payment_records;

ALTER TABLE orders 
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS payment_status,
DROP COLUMN IF EXISTS payment_intent_id;


/*


payment_records sem CASCADE
sqlDROP TABLE IF EXISTS payment_records;
Se payment_records for referenciada por alguma tabela futura, vai falhar. Mantenha o padrão:
sqlDROP TABLE IF EXISTS payment_records CASCADE;




*/