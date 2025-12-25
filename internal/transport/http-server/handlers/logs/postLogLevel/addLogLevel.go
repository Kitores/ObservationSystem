package postLogLevel

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type LoglevelAdder interface {
	PostNewLogLevel(levelName string, severity int, description string) error
}

type request struct {
	LevelName   string `json:"levelName"`
	Severity    int    `json:"severity"`
	Description string `json:"description"`
}

func New(logLevelAdder LoglevelAdder) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(request)
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		err := logLevelAdder.PostNewLogLevel(req.LevelName, req.Severity, req.Description)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "error adding the log level"})
		}
		return c.JSON(http.StatusOK, map[string]string{"levelName": req.LevelName})
	}
}
