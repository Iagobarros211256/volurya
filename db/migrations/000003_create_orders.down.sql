DROP TABLE IF EXISTS orders;

/*

 Sem CASCADE
orders é referenciada por order_items e possivelmente payments. Sem CASCADE vai falhar:
sqlDROP TABLE IF EXISTS orders CASCADE;
Todas as down migrations do projeto têm esse problema. Nas próximas vou assumir que é recorrente e focar em problemas novos

*/