package login

import (
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/authorization/jwtTypes"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type LoginSaver interface {
	UserExists(username string) (jwtTypes.User, error)
}

func New(loginSaver LoginSaver, jwtSecret string) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req jwtTypes.LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{"error": "Bad request"})
		}

		// Ищем пользователя
		user, err := loginSaver.UserExists(req.Name)
		if err != nil {
			return c.JSON(401, map[string]string{"error": "Invalid credentials"})
		}

		// Проверяем пароль
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			return c.JSON(401, map[string]string{"error": "Invalid credentials"})
		}
		// Создаем JWT токен
		token, expTime, err := jwtTypes.CreateJWTToken(user.ID, user.Name, jwtSecret)
		if err != nil {
			return c.JSON(500, map[string]string{"error": "Cannot create token"})
		}

		return c.JSON(200, map[string]interface{}{
			"token":      token,
			"expires_at": expTime.Format(time.RFC3339),
			"user": map[string]interface{}{
				"id":   user.ID,
				"name": user.Name,
			},
		})
	}
}
