package postEnvironment

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type EnvironmentAdder interface {
	PostNewEnvironment(envName string, description string) error
}

type request struct {
	EnvName     string `json:"name"`
	Description string `json:"description"`
}

func New(environmentAdder EnvironmentAdder) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(request)
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		err := environmentAdder.PostNewEnvironment(req.EnvName, req.Description)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "error adding the environment"})
		}
		return c.JSON(http.StatusOK, map[string]string{"envName": req.EnvName})
	}
}
