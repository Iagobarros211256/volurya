package jobs

import (
	"context"
	"log/slog"
	"time"
)

// RefreshTokenCleaner define a interface necessária para o cleanup job
type RefreshTokenCleaner interface {
	DeleteExpired() (int64, error)
}

// StartTokenCleanup inicia um job periódico que remove refresh tokens expirados e revogados.
// Roda a cada interval e para quando o context for cancelado.
func StartTokenCleanup(ctx context.Context, repo RefreshTokenCleaner, interval time.Duration) {
	go func() {
		slog.Info("token cleanup job started", "interval", interval)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Roda uma vez imediatamente no startup
		runCleanup(repo)

		for {
			select {
			case <-ctx.Done():
				slog.Info("token cleanup job stopped")
				return
			case <-ticker.C:
				runCleanup(repo)
			}
		}
	}()
}

func runCleanup(repo RefreshTokenCleaner) {
	deleted, err := repo.DeleteExpired()
	if err != nil {
		slog.Error("token cleanup failed", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("token cleanup completed", "deleted", deleted)
	}
}
