package registration

import (
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/authorization/jwtTypes"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Registrator interface {
	NewUser(username, hashedPassword string) (int, error)
}

func New(registrator Registrator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req jwtTypes.LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{"error": "Bad request"})
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(500, map[string]string{"error": "Server error"})
		}

		// Сохраняем в БД
		userID, err := registrator.NewUser(req.Name, string(hashedPassword))

		if err != nil {
			return c.JSON(400, map[string]string{"error": "User already exists"})
		}

		return c.JSON(200, map[string]interface{}{
			"message": "User registered successfully",
			"user_id": userID,
		})
	}
}
