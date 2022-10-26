package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
)

func (crtl *Controller) Health(c echo.Context) error {

	var unhealthyMessage string

	dbCtx, dbCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer dbCancel()
	if err := crtl.db.PingContext(dbCtx); err != nil {
		unhealthyMessage += "unable to connect to postgres"
	}

	rdbCtx, rdbCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer rdbCancel()
	if err := crtl.rdb.Ping(rdbCtx).Err(); err != nil {
		if unhealthyMessage != "" {
			unhealthyMessage += "; unable to connect to redis"
		} else {
			unhealthyMessage += "unable to connect to redis"
		}
	}

	res := make(map[string]interface{})
	if unhealthyMessage != "" {
		res["code"] = http.StatusInternalServerError
		res["message"] = unhealthyMessage
		return c.JSON(http.StatusInternalServerError, res)
	}

	res["code"] = http.StatusOK
	res["message"] = "healthy"
	return c.JSON(http.StatusOK, res)
}
