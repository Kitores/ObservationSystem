package http_server

import (
	"errors"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/storage/postgre/methods"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/environments/postEnvironment"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/hosts/getHosts"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/login"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/logs/getLogs"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/logs/postLogLevel"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/registration"
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/handlers/services/getServices"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"os"
)

type JWTClaims struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
}

func StartHttpServer(pg *methods.PostgreSqlx) {

	err := godotenv.Load("token.env")
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}
	jwtSecret := os.Getenv("JWT_TOKEN")
	e := echo.New()

	e.POST("/login", login.New(pg, jwtSecret))
	e.POST("/register", registration.New(pg))

	jwtSecretBytes := []byte(jwtSecret)

	jwtConfig := echojwt.Config{
		SigningKey: jwtSecretBytes,
	}

	protected := e.Group("/api")
	protected.Use(echojwt.WithConfig(jwtConfig))

	protected.GET("/getLogs", getLogs.New(pg))
	protected.GET("/getHosts", getHosts.New(pg))
	protected.GET("/getServices", getServices.New(pg))

	protected.POST("/postNewLevel", postLogLevel.New(pg))
	protected.POST("/postNewEnv", postEnvironment.New(pg))

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
