package booking

import (
	"context"
	"log"
	"time"
)

func StartExpiryWorker(ctx context.Context, service *Service) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := service.ExpirePendingBookings(ctx); err != nil {
					log.Printf("booking expiry worker: %v", err)
				}
			}
		}
	}()
}
