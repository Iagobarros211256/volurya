package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init() {
	env := os.Getenv("APP_ENV")

	var handler slog.Handler

	if env == "production" {
		// JSON estruturado em produção
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Texto legível em desenvolvimento
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}

/*
Log pode ser usado antes de Init() ser chamado
govar Log *slog.Logger
Se qualquer pacote chamar logger.Log.Info(...) antes de main() chamar logger.Init(), vai panic com nil pointer. Inicialize com um default seguro:
govar Log = slog.Default()

🟡 Sem suporte a LevelWarn configurável
O nível é hardcoded por ambiente. Em produção pode ser útil aumentar para LevelWarn temporariamente para reduzir volume de logs. Torne configurável via env:
golevelStr := os.Getenv("LOG_LEVEL")
level := slog.LevelInfo
switch levelStr {
case "debug": level = slog.LevelDebug
case "warn":  level = slog.LevelWarn
case "error": level = slog.LevelError
}

🟡 Sem AddSource em desenvolvimento
Para debugging local, logar o arquivo e linha de origem é muito útil:
gohandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level:     slog.LevelDebug,
    AddSource: true,
})

🟢 APP_ENV verificado aqui e ENV em outros lugares
payment_controller.go usa APP_ENV, conn.go e outros usam ENV. Duas variáveis de ambiente para o mesmo conceito. Padronize uma — APP_ENV ou ENV — em todo o projeto

*/
