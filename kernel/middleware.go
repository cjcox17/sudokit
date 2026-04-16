package kernel

import (
	"github.com/gofiber/fiber/v2"
)

func ServiceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		k := App()
		if k == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "service unavailable",
				"code":  "KERNEL_NOT_INITIALIZED",
			})
		}

		ctx := WithServices(c.UserContext(), k)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

func ServiceMiddlewareWithKernel(k *Kernel) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := WithServices(c.UserContext(), k)
		c.SetUserContext(ctx)
		return c.Next()
	}
}
