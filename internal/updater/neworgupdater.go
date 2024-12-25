package updater

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github-copilot-invite/internal/handlers"

	"github.com/gin-gonic/gin"
)

type OrgsTrigger struct {
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
}

func NewOrgsTrigger(h *handlers.Handler) *OrgsTrigger {
	// create a new gin engine
	ctx, cancel := context.WithCancel(context.Background())
	t := &OrgsTrigger{
		ticker: time.NewTicker(10 * time.Second),
		// ticker: time.NewTicker(1 * time.Hour),
		ctx:    ctx,
		cancel: cancel,
	}

	// create a new gin engine
	// engine := gin.Default()

	// create a new handler

	c, _ := gin.CreateTestContext(nil)
	// c, _ := engine.CreateTestContext

	go func() {
		for {
			select {
			case <-t.ticker.C:
				// h.ListOrganizations(c)
				h.HealthCheck(c)
			case <-t.ctx.Done():
				t.ticker.Stop()
				return
			}
		}
	}()

	// Catch SIGINT and SIGTERM signals and cancel the context
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		t.cancel()
	}()

	return t
}
