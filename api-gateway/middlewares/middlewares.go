package middlewares

import (
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	Auth auth.AuthServiceClient
}

func NewHandler(authHandler auth.AuthServiceClient) *Handler {
	return &Handler{
		Auth: authHandler,
	}
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func (h *Handler) Authenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, &ErrorMessage{
				Message: "Unauthorized",
			})
		}

		checkRes, err := h.Auth.Check(c.Request().Context(), &auth.CheckRequest{
			Token: token,
		})

		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusUnauthorized, &ErrorMessage{
				Message: "Unauthorized",
			})
		}

		c.Set("user", checkRes.User)
		c.Set("token", checkRes.Token)
		return next(c)
	}
}
