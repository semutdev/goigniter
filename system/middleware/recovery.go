package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/semutdev/goigniter/system/core"
)

// Recovery returns a middleware that recovers from panics.
func Recovery() core.Middleware {
	return RecoveryWithConfig(RecoveryConfig{})
}

// RecoveryConfig holds configuration for the recovery middleware.
type RecoveryConfig struct {
	StackSize         int
	DisableStackAll   bool
	DisablePrintStack bool
	LogFunc           func(c *core.Context, err any, stack []byte)
	// HideErrorDetails hides error details in production (recommended)
	HideErrorDetails bool
	// ProductionMessage is the message shown when HideErrorDetails is true
	ProductionMessage string
}

// RecoveryWithConfig returns a Recovery middleware with custom config.
func RecoveryWithConfig(config RecoveryConfig) core.Middleware {
	if config.StackSize == 0 {
		config.StackSize = 4 << 10
	}
	if config.ProductionMessage == "" {
		config.ProductionMessage = "Internal Server Error"
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			defer func() {
				if r := recover(); r != nil {
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					stack = stack[:length]

					if config.LogFunc != nil {
						config.LogFunc(c, r, stack)
					} else if !config.DisablePrintStack {
						log.Printf("[PANIC RECOVER] %v\n%s", r, stack)
					}

					// Hide error details in production
					if config.HideErrorDetails {
						c.String(http.StatusInternalServerError, config.ProductionMessage)
					} else {
						c.String(http.StatusInternalServerError,
							fmt.Sprintf("Internal Server Error: %v", r))
					}
				}
			}()

			return next(c)
		}
	}
}
