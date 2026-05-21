DROP TABLE IF EXISTS refresh_tokens;

/*



7:55 AM
Mesmo problema — sem CASCADE. Mas refresh_tokens provavelmente não é referenciada por outras tabelas, então nesse caso específico o DROP provavelmente funcionaria sem ele.

Ainda assim, mantenha consistência:

sql
DROP TABLE IF EXISTS refresh_tokens CASCADE;

*/