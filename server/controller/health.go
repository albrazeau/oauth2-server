package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (crtl *Controller) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"code": 200, "status": "healthy"})
}
