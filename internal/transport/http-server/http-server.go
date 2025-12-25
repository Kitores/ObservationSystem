package http_server

import (
	"errors"
	"github.com/Kitores/ObservationSystem/internal/storage/postgre/methods"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/environments/postEnvironment"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/hosts/getHosts"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/logs/getLogs"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/logs/postLogLevel"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/services/getServices"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

func StartHttpServer(pg *methods.PostgreSqlx) {
	e := echo.New()

	e.GET("/getLogs", getLogs.New(pg))
	e.GET("/getHosts", getHosts.New(pg))
	e.GET("/getServices", getServices.New(pg))

	e.POST("/postNewLevel", postLogLevel.New(pg))
	e.POST("/postNewEnv", postEnvironment.New(pg))

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
