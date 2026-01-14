package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/semutdev/goigniter/system/core"
)

// Logger returns a middleware that logs HTTP requests.
func Logger() core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)
			log.Printf("[%s] %s %s %v",
				c.Method(),
				c.Path(),
				latency,
				err,
			)
			return err
		}
	}
}

// LoggerConfig holds configuration for the logger middleware.
type LoggerConfig struct {
	Format    string
	SkipPaths []string
}

// LoggerWithConfig returns a Logger middleware with custom config.
func LoggerWithConfig(config LoggerConfig) core.Middleware {
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			if skipMap[c.Path()] {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			if config.Format != "" {
				log.Printf(config.Format, c.Method(), c.Path(), latency)
			} else {
				log.Printf("[%s] %s %s",
					c.Method(),
					c.Path(),
					latency,
				)
			}

			return err
		}
	}
}

// ColorLogger returns a middleware that logs with colors.
func ColorLogger() core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			methodColor := methodToColor(c.Method())
			resetColor := "\033[0m"

			fmt.Printf("%s[%s]%s %s %v\n",
				methodColor,
				c.Method(),
				resetColor,
				c.Path(),
				latency,
			)

			return err
		}
	}
}

func methodToColor(method string) string {
	switch method {
	case "GET":
		return "\033[32m"
	case "POST":
		return "\033[34m"
	case "PUT":
		return "\033[33m"
	case "DELETE":
		return "\033[31m"
	case "PATCH":
		return "\033[36m"
	default:
		return "\033[0m"
	}
}
