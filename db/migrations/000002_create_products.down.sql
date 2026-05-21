DROP TABLE IF EXISTS products;

/*
 Sem CASCADE
products é referenciada por order_items, cart_items e possivelmente outras tabelas. Sem CASCADE esse DROP vai falhar se houver dependências:
sqlDROP TABLE IF EXISTS products CASCADE;

*/