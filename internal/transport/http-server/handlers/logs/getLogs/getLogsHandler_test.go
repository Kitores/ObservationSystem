package getLogs

import (
	"encoding/json"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// MockLogsGetter имитирует интерфейс LogsGetter
type MockLogsGetter struct {
	mock.Mock
}

func (m *MockLogsGetter) GetRecentLogsByInterval(filterType string, filterValue interface{}, value int, unit string) ([]entity.LogEntity, error) {
	args := m.Called(filterType, filterValue, value, unit)
	return args.Get(0).([]entity.LogEntity), args.Error(1)
}

func TestGetLogsHandler(t *testing.T) {
	// Тестовые данные
	mockLogs := []entity.LogEntity{
		{
			ID:        1,
			Message:   "Test log message",
			Timestamp: time.Now(),
		},
	}

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*MockLogsGetter)
		expectedStatus int
	}{
		{
			name:        "successful request with ip filter",
			requestBody: `{"filter_type": "ip", "filter_value": "192.168.1.1", "value": 24, "unit": "hours"}`,
			setupMock: func(m *MockLogsGetter) {
				m.On("GetRecentLogsByInterval", "ip", "192.168.1.1", 24, "hours").Return(mockLogs, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful request with host_id filter (number)",
			requestBody: `{"filter_type": "host_id", "filter_value": 123, "value": 1, "unit": "day"}`,
			setupMock: func(m *MockLogsGetter) {
				m.On("GetRecentLogsByInterval", "host_id", int64(123), 1, "day").Return(mockLogs, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful request with host_id filter (string)",
			requestBody: `{"filter_type": "host_id", "filter_value": "123", "value": 1, "unit": "day"}`,
			setupMock: func(m *MockLogsGetter) {
				m.On("GetRecentLogsByInterval", "host_id", int64(123), 1, "day").Return(mockLogs, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful request with host_name filter",
			requestBody: `{"filter_type": "host_name", "filter_value": "server-01", "value": 30, "unit": "minutes"}`,
			setupMock: func(m *MockLogsGetter) {
				m.On("GetRecentLogsByInterval", "host_name", "server-01", 30, "minutes").Return(mockLogs, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid filter type",
			requestBody:    `{"filter_type": "invalid", "filter_value": "test", "value": 1, "unit": "hour"}`,
			setupMock:      func(m *MockLogsGetter) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			requestBody:    `{bad json}`,
			setupMock:      func(m *MockLogsGetter) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid host_id value",
			requestBody:    `{"filter_type": "host_id", "filter_value": "not-a-number", "value": 1, "unit": "day"}`,
			setupMock:      func(m *MockLogsGetter) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid filter value for string type",
			requestBody:    `{"filter_type": "ip", "filter_value": 123, "value": 1, "unit": "day"}`,
			setupMock:      func(m *MockLogsGetter) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "database error",
			requestBody: `{"filter_type": "ip", "filter_value": "192.168.1.1", "value": 24, "unit": "hours"}`,
			setupMock: func(m *MockLogsGetter) {
				m.On("GetRecentLogsByInterval", "ip", "192.168.1.1", 24, "hours").Return([]entity.LogEntity{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "empty request body",
			requestBody:    "",
			setupMock:      func(m *MockLogsGetter) {},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/getLogs", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Создаем и настраиваем мок
			mockGetter := new(MockLogsGetter)
			tt.setupMock(mockGetter)

			handler := New(mockGetter)

			err := handler(c)

			if tt.expectedStatus >= 400 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var logs []entity.LogEntity
				err := json.Unmarshal(rec.Body.Bytes(), &logs)
				assert.NoError(t, err)
				assert.Len(t, logs, 1)
			}
			mockGetter.AssertExpectations(t)
		})
	}
}
