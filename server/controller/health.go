package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
)

func (crtl *Controller) Health(c echo.Context) error {

	dbCtx, dbCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer dbCancel()
	if err := crtl.db.PingContext(dbCtx); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "message": "healthy"})
}
