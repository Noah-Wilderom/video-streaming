package handlers

import (
	"context"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	loginCtx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	loginResponse, err := h.Auth.Login(loginCtx, &auth.LoginRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		fmt.Printf("%+v\n", err)
		if strings.Contains(err.Error(), "credentials") {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": "invalid credentials",
			})
		}
		return err
	}

	return c.JSON(http.StatusOK, loginResponse.User)
}

func (h *Handler) Register(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	registerCtx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	registerResponse, err := h.Auth.Register(registerCtx, &auth.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, registerResponse)
}

func (h *Handler) Check(c echo.Context) error {
	checkCtx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	checkResponse, err := h.Auth.Check(checkCtx, &auth.CheckRequest{
		Token: c.Get("token").(string),
	})

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, checkResponse)
}
