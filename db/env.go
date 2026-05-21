package db

import "os"

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

/*

🟡 Duplicação com o pacote config
O pacote config provavelmente tem função idêntica. Ter getEnv em dois pacotes diferentes (db e config) fragmenta a lógica de configuração. O ideal seria um único utilitário no pacote config usado por todos.

🟡 Não exportada — limita reuso
Como é getEnv (minúsculo), só é acessível dentro do pacote db. Se outros pacotes precisarem do mesmo padrão, reimplementam. Se mover para config:
go// config/env.go
func GetEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

🟢 Não distingue variável ausente de variável vazia
goif value := os.Getenv(key); value != "" {
os.Getenv("KEY") retorna "" tanto para variável não definida quanto para variável definida como string vazia (KEY=""). Se precisar distinguir os dois casos use os.LookupEnv:
gofunc getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}
Para a maioria dos casos do projeto isso não importa, mas para configurações críticas como DATABASE_URL a distinção é relevante.

*/
