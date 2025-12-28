package login

import (
	"github.com/Kitores/ObservationSystem/internal/transport/http-server/authorization/jwtTypes"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockLoginSaver имитирует интерфейс LoginSaver
type MockLoginSaver struct {
	mock.Mock
}

func (m *MockLoginSaver) UserExists(username string) (jwtTypes.User, error) {
	args := m.Called(username)
	return args.Get(0).(jwtTypes.User), args.Error(1)
}

// TestLoginHandler тестирует обработчик авторизации
func TestLoginHandler(t *testing.T) {
	// Тестовые данные
	testJWTSecret := "test-secret-key-1234567890"
	testUser := jwtTypes.User{
		ID:       1,
		Name:     "testuser",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye0nG9q8M2p5JQ8J3qkZJ7sF4J9vY3LmO", // bcrypt hash для "password123"
	}

	// Установим фиксированное время для предсказуемых тестов
	//fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(mock *MockLoginSaver)
		expectedStatus int
		expectedBody   string
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:        "successful login",
			requestBody: `{"name": "testuser", "password": "password123"}`,
			setupMock: func(mock *MockLoginSaver) {
				mock.On("UserExists", "testuser").Return(testUser, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "token")
				assert.Contains(t, body, "expires_at")
				assert.Contains(t, body, "user")
				assert.Contains(t, body, "testuser")

				// Проверяем, что токен валиден
				tokenString := extractTokenFromJSON(body)
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(testJWTSecret), nil
				})
				assert.NoError(t, err)
				assert.True(t, token.Valid)
			},
		},
		{
			name:        "invalid credentials - wrong password",
			requestBody: `{"name": "testuser", "password": "wrongpassword"}`,
			setupMock: func(mock *MockLoginSaver) {
				mock.On("UserExists", "testuser").Return(testUser, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid credentials"}`,
		},
		{
			name:        "user not found",
			requestBody: `{"name": "nonexistent", "password": "password123"}`,
			setupMock: func(mock *MockLoginSaver) {
				mock.On("UserExists", "nonexistent").Return(jwtTypes.User{}, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid credentials"}`,
		},
		{
			name:        "invalid json",
			requestBody: `{"name": "testuser", "password": "password123",}`,
			setupMock: func(mock *MockLoginSaver) {
				// Мок не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Bad request"}`,
		},
		{
			name:        "empty request body",
			requestBody: "",
			setupMock: func(mock *MockLoginSaver) {
				// Мок не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Bad request"}`,
		},
		{
			name:        "missing username",
			requestBody: `{"password": "password123"}`,
			setupMock: func(mock *MockLoginSaver) {
				// Мок не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Bad request"}`,
		},
		{
			name:        "missing password",
			requestBody: `{"name": "testuser"}`,
			setupMock: func(mock *MockLoginSaver) {
				// Мок не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Bad request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем Echo контекст
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Создаем и настраиваем мок
			mockLoginSaver := new(MockLoginSaver)
			tt.setupMock(mockLoginSaver)

			// Создаем обработчик
			handler := New(mockLoginSaver, testJWTSecret)

			// Вызываем обработчик
			err := handler(c)

			// Проверяем результат
			if tt.expectedStatus >= 400 {
				assert.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.expectedStatus, httpErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Проверяем тело ответа
			body := strings.TrimSpace(rec.Body.String())
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, body)
			}

			// Дополнительные проверки если есть
			if tt.checkResponse != nil {
				tt.checkResponse(t, body)
			}

			// Проверяем, что все ожидания мока выполнены
			mockLoginSaver.AssertExpectations(t)
		})
	}
}

// TestLoginHandler_TokenCreationError тестирует ошибку создания токена
func TestLoginHandler_TokenCreationError(t *testing.T) {
	// Пустой секрет вызовет ошибку при создании токена
	emptySecret := ""
	testUser := jwtTypes.User{
		ID:       1,
		Name:     "testuser",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye0nG9q8M2p5JQ8J3qkZJ7sF4J9vY3LmO", // bcrypt hash для "password123"
	}

	// Создаем Echo контекст
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login",
		strings.NewReader(`{"name": "testuser", "password": "password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Создаем мок
	mockLoginSaver := new(MockLoginSaver)
	mockLoginSaver.On("UserExists", "testuser").Return(testUser, nil)

	// Создаем обработчик с пустым секретом
	handler := New(mockLoginSaver, emptySecret)

	// Вызываем обработчик
	err := handler(c)

	// Проверяем результат
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	if ok {
		assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	}

	// Проверяем тело ответа
	body := strings.TrimSpace(rec.Body.String())
	assert.Contains(t, body, "Cannot create token")

	// Проверяем мок
	mockLoginSaver.AssertExpectations(t)
}

// TestNewLoginHandler тестирует создание обработчика
func TestNewLoginHandler(t *testing.T) {
	mockLoginSaver := new(MockLoginSaver)
	jwtSecret := "test-secret"

	handler := New(mockLoginSaver, jwtSecret)

	assert.NotNil(t, handler)
	assert.IsType(t, echo.HandlerFunc(nil), handler)
}

// TestLoginHandler_ContentType проверяет обработку разных Content-Type
func TestLoginHandler_ContentType(t *testing.T) {
	testUser := jwtTypes.User{
		ID:       1,
		Name:     "testuser",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye0nG9q8M2p5JQ8J3qkZJ7sF4J9vY3LmO",
	}

	tests := []struct {
		name           string
		contentType    string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "valid json content type",
			contentType:    echo.MIMEApplicationJSON,
			requestBody:    `{"name": "testuser", "password": "password123"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid content type",
			contentType:    echo.MIMETextPlain,
			requestBody:    `{"name": "testuser", "password": "password123"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no content type",
			contentType:    "",
			requestBody:    `{"name": "testuser", "password": "password123"}`,
			expectedStatus: http.StatusOK, // Echo обрабатывает и без заголовка
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set(echo.HeaderContentType, tt.contentType)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockLoginSaver := new(MockLoginSaver)
			if tt.expectedStatus == http.StatusOK {
				mockLoginSaver.On("UserExists", "testuser").Return(testUser, nil)
			}

			handler := New(mockLoginSaver, "test-secret")
			err := handler(c)

			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockLoginSaver.AssertExpectations(t)
		})
	}
}

// TestLoginHandler_BcryptError тестирует обработку ошибок bcrypt
func TestLoginHandler_BcryptError(t *testing.T) {
	// Пользователь с невалидным хешем пароля
	testUser := jwtTypes.User{
		ID:       1,
		Name:     "testuser",
		Password: "invalid-hash", // Невалидный bcrypt хеш
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login",
		strings.NewReader(`{"name": "testuser", "password": "password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLoginSaver := new(MockLoginSaver)
	mockLoginSaver.On("UserExists", "testuser").Return(testUser, nil)

	handler := New(mockLoginSaver, "test-secret")
	err := handler(c)

	// Должна быть ошибка сравнения пароля
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid credentials")

	mockLoginSaver.AssertExpectations(t)
}

// TestLoginHandler_EmptyPassword тестирует пустой пароль
func TestLoginHandler_EmptyPassword(t *testing.T) {
	testUser := jwtTypes.User{
		ID:       1,
		Name:     "testuser",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye0nG9q8M2p5JQ8J3qkZJ7sF4J9vY3LmO",
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login",
		strings.NewReader(`{"name": "testuser", "password": ""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLoginSaver := new(MockLoginSaver)
	mockLoginSaver.On("UserExists", "testuser").Return(testUser, nil)

	handler := New(mockLoginSaver, "test-secret")
	err := handler(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid credentials")

	mockLoginSaver.AssertExpectations(t)
}

// Вспомогательная функция для извлечения токена из JSON
func extractTokenFromJSON(jsonStr string) string {
	// Простая реализация для тестов
	// В реальном проекте лучше использовать json.Unmarshal
	start := strings.Index(jsonStr, `"token":"`) + 9
	end := strings.Index(jsonStr[start:], `"`)
	if start >= 9 && end >= 0 {
		return jsonStr[start : start+end]
	}
	return ""
}

// TestLoginHandler_UserDetails проверяет детали пользователя в ответе
func TestLoginHandler_UserDetails(t *testing.T) {
	testUser := jwtTypes.User{
		ID:       42,
		Name:     "john_doe",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye0nG9q8M2p5JQ8J3qkZJ7sF4J9vY3LmO",
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login",
		strings.NewReader(`{"name": "john_doe", "password": "password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLoginSaver := new(MockLoginSaver)
	mockLoginSaver.On("UserExists", "john_doe").Return(testUser, nil)

	handler := New(mockLoginSaver, "test-secret")
	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.Contains(t, body, `"id":42`)
	assert.Contains(t, body, `"name":"john_doe"`)
	assert.Contains(t, body, `"user":{`)
	assert.Contains(t, body, `"expires_at"`)

	mockLoginSaver.AssertExpectations(t)
}

// TestLoginHandler_MethodNotAllowed проверяет другие HTTP методы
func TestLoginHandler_MethodNotAllowed(t *testing.T) {
	e := echo.New()

	methods := []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/login", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockLoginSaver := new(MockLoginSaver)
			handler := New(mockLoginSaver, "test-secret")

			// Для методов отличных от POST Echo может вернуть 405 или ошибку
			err := handler(c)

			if method != http.MethodPost {
				// Ожидаем ошибку, так как метод не POST
				assert.Error(t, err)
			}
		})
	}
}
