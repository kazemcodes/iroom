package handler

import "github.com/labstack/echo/v4"

func getUserID(c echo.Context) (int64, bool) {
	id, ok := c.Get("user_id").(int64)
	return id, ok
}

func getUserRole(c echo.Context) string {
	role, _ := c.Get("role").(string)
	return role
}
