// Package middleware provides a fiber middleware for logging requests.
package middleware

import (
	"lib/logger"

	"github.com/gofiber/fiber/v2"
)

// Middleware returns a fiber handler using the global log instance.
func Middleware(verbose bool) fiber.Handler {
	sublog := logger.Get()

	return func(c *fiber.Ctx) error {
		chainErr := c.Next()

		msg := "Request"
		if chainErr != nil {
			msg = chainErr.Error()
		}

		code := c.Response().StatusCode()

		log := sublog.With().
			Int("status", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Str("user-agent", c.Get(fiber.HeaderUserAgent)).
			Logger()

		switch {
		case code >= 200 && code < 300:
			if verbose {
				log.Info().Msg(msg)
			} else {
				log.Debug().Msg(msg)
			}
		case code >= 300 && code < 400:
			log.Warn().Msg(msg)
		default:
			log.Error().Msg(msg)
		}
		return chainErr
	}
}
