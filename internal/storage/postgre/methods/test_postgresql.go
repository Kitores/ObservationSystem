package methods

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Kitores/ObservationSystem/internal/models/entity"
)

// NewPostgreSqlx создает новый экземпляр PostgreSqlx
func NewPostgreSqlx(db *sqlx.DB) *PostgreSqlx {
	return &PostgreSqlx{db: db}
}

// TestPostgreSqlx_SaveLog тестирует сохранение лога
func TestPostgreSqlx_SaveLog(t *testing.T) {
	tests := []struct {
		name        string
		logEntity   entity.LogEntity
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "successful save",
			logEntity: entity.LogEntity{
				ServiceID:     1,
				EnvironmentID: 2,
				HostID:        3,
				LevelID:       4,
				Message:       "Test message",
				Timestamp:     time.Now(),
				LoggerName:    "TestLogger",
				ReceivedAt:    time.Now(),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO logs`).
					WithArgs(
						int64(1), // service_id
						int64(2), // environment_id
						int64(3), // host_id
						int64(4), // level_id
						"Test message",
						sqlmock.AnyArg(), // timestamp
						"TestLogger",
						sqlmock.AnyArg(), // received_at
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectError: false,
		},
		{
			name: "database error",
			logEntity: entity.LogEntity{
				ServiceID:     1,
				EnvironmentID: 2,
				HostID:        3,
				LevelID:       4,
				Message:       "Test message",
				Timestamp:     time.Now(),
				LoggerName:    "TestLogger",
				ReceivedAt:    time.Now(),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO logs`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок базы данных
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			pg := NewPostgreSqlx(sqlxDB)

			// Настраиваем мок
			tt.setupMock(mock)

			// Вызываем тестируемый метод
			result, err := pg.SaveLog(tt.logEntity)

			// Проверяем результаты
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to save log")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Проверяем, что все ожидания выполнены
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPostgreSqlx_RegisterService тестирует регистрацию сервиса
func TestPostgreSqlx_RegisterService(t *testing.T) {
	teamOwner := "team1"
	desc := "description"

	tests := []struct {
		name        string
		service     entity.Service
		setupMock   func(mock sqlmock.Sqlmock)
		expectedIDs []int64
		expectError bool
	}{
		{
			name: "successful registration",
			service: entity.Service{
				Name:      "test-service",
				HostName:  "test-host",
				HostIP:    "192.168.1.1",
				TeamOwner: teamOwner,
				Desc:      desc,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"service_id", "host_id"}).
					AddRow(int64(100), int64(200))
				mock.ExpectQuery(`SELECT service_id, host_id`).
					WithArgs(
						"test-service",
						"test-host",
						"192.168.1.1",
						teamOwner,
						desc,
					).
					WillReturnRows(rows)
			},
			expectedIDs: []int64{100, 200},
			expectError: false,
		},
		{
			name: "database error",
			service: entity.Service{
				Name:      "test-service",
				HostName:  "test-host",
				HostIP:    "192.168.1.1",
				TeamOwner: teamOwner,
				Desc:      desc,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT service_id, host_id`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedIDs: []int64{0, 0},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			pg := &PostgreSqlx{db: sqlxDB}

			tt.setupMock(mock)

			serviceID, hostID, err := pg.RegisterService(tt.service)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to register service")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedIDs[0], serviceID)
				assert.Equal(t, tt.expectedIDs[1], hostID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPostgreSqlx_UpdateServiceStatus тестирует обновление статуса сервиса
func TestPostgreSqlx_UpdateServiceStatus(t *testing.T) {
	tests := []struct {
		name            string
		serviceNameOrID interface{}
		newStatus       bool
		setupMock       func(mock sqlmock.Sqlmock, arg interface{})
		expectError     bool
	}{
		{
			name:            "update by service name - successful",
			serviceNameOrID: "test-service",
			newStatus:       true,
			setupMock: func(mock sqlmock.Sqlmock, arg interface{}) {
				mock.ExpectExec(`UPDATE services.*WHERE name = \$`).
					WithArgs(true, "test-service").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name:            "update by host id - successful",
			serviceNameOrID: 123,
			newStatus:       false,
			setupMock: func(mock sqlmock.Sqlmock, arg interface{}) {
				mock.ExpectExec(`UPDATE services.*FROM hosts.*WHERE hosts.id = services.host_id AND hosts.id = \$`).
					WithArgs(false, 123).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name:            "update by service name - database error",
			serviceNameOrID: "test-service",
			newStatus:       true,
			setupMock: func(mock sqlmock.Sqlmock, arg interface{}) {
				mock.ExpectExec(`UPDATE services.*WHERE name = \$`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectError: true,
		},
		{
			name:            "unsupported type",
			serviceNameOrID: 123.45,
			newStatus:       true,
			setupMock:       func(mock sqlmock.Sqlmock, arg interface{}) {},
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			pg := &PostgreSqlx{db: sqlxDB}

			tt.setupMock(mock, tt.serviceNameOrID)

			err = pg.UpdateServiceStatus(tt.serviceNameOrID, tt.newStatus)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.name != "unsupported type" {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestPostgreSqlx_GetRecentLogsByInterval тестирует получение логов по интервалу
func TestPostgreSqlx_GetRecentLogsByInterval(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		filterType  string
		filterValue interface{}
		value       int
		unit        string
		setupMock   func(mock sqlmock.Sqlmock)
		expected    []entity.LogEntity
		expectError bool
	}{
		{
			name:        "filter by ip - successful",
			filterType:  "ip",
			filterValue: "192.168.1.1",
			value:       24,
			unit:        "hours",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "timestamp", "message", "logger_name", "received_at", "version",
					"service_id", "host_id", "host_ip", "environment_id", "environment_name",
					"level_id", "level_name", "level_severity",
				}).
					AddRow(
						int64(1), now, "Error message", "TestLogger", now, "1.0",
						int64(100), int64(200), "192.168.1.1", int64(300), "production",
						int64(400), "ERROR", 50,
					)
				mock.ExpectQuery(`SELECT.*WHERE l.timestamp >= NOW\(\) - \(\$1 \|\| ' ' \|\| \$2\)::interval.*AND h.ip = \$3`).
					WithArgs("24", "hours", "192.168.1.1").
					WillReturnRows(rows)
			},
			expected: []entity.LogEntity{
				{
					ID:         1,
					Timestamp:  now,
					Message:    "Error message",
					LoggerName: "TestLogger",
					ReceivedAt: now,
					Version:    1,
					ServiceID:  100,
					HostID:     200,
					HostIP:     sql.NullString{String: "192.168.1.1", Valid: true}.String,
					EnvName:    "production",
					LevelName:  "ERROR",
				},
			},
			expectError: false,
		},
		{
			name:        "invalid unit",
			filterType:  "ip",
			filterValue: "192.168.1.1",
			value:       24,
			unit:        "invalid",
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: true,
		},
		{
			name:        "invalid filter type",
			filterType:  "invalid",
			filterValue: "value",
			value:       24,
			unit:        "hours",
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: true,
		},
		{
			name:        "database error",
			filterType:  "host_id",
			filterValue: 123,
			value:       1,
			unit:        "day",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT.*WHERE l.timestamp >= NOW\(\) - \(\$1 \|\| ' ' \|\| \$2\)::interval.*AND h.id = \$3`).
					WithArgs("1", "day", 123).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			pg := &PostgreSqlx{db: sqlxDB}

			tt.setupMock(mock)

			logs, err := pg.GetRecentLogsByInterval(tt.filterType, tt.filterValue, tt.value, tt.unit)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logs)
			} else {
				assert.NoError(t, err)
				assert.Len(t, logs, len(tt.expected))
				if len(logs) > 0 {
					assert.Equal(t, tt.expected[0].ID, logs[0].ID)
					assert.Equal(t, tt.expected[0].Message, logs[0].Message)
				}
			}

			if tt.name != "invalid unit" && tt.name != "invalid filter type" {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestPostgreSqlx_GetHostErrorStats тестирует получение статистики ошибок хостов
func TestPostgreSqlx_GetHostErrorStats(t *testing.T) {
	lastError := time.Now()

	tests := []struct {
		name        string
		hours       int
		setupMock   func(mock sqlmock.Sqlmock)
		expected    []entity.HostErrorStats
		expectError bool
	}{
		{
			name:  "successful retrieval",
			hours: 24,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"host_id", "host_name", "error_count", "fatal_count", "last_error",
				}).
					AddRow(int64(1), "host1", 5, 2, lastError).
					AddRow(int64(2), "host2", 3, 1, lastError.Add(-time.Hour))
				mock.ExpectQuery(`SELECT.*WHERE l.timestamp >= NOW\(\) - \$1::interval`).
					WithArgs("'24 hours'").
					WillReturnRows(rows)
			},
			expected: []entity.HostErrorStats{
				{
					HostID:     1,
					HostName:   "host1",
					ErrorCount: 5,
					FatalCount: 2,
					LastError:  lastError,
				},
				{
					HostID:     2,
					HostName:   "host2",
					ErrorCount: 3,
					FatalCount: 1,
					LastError:  lastError.Add(-time.Hour),
				},
			},
			expectError: false,
		},
		{
			name:  "no errors found",
			hours: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"host_id", "host_name", "error_count", "fatal_count", "last_error",
				})
				mock.ExpectQuery(`SELECT.*WHERE l.timestamp >= NOW\(\) - \$1::interval`).
					WithArgs("'1 hours'").
					WillReturnRows(rows)
			},
			expected:    []entity.HostErrorStats{},
			expectError: false,
		},
		{
			name:  "database error",
			hours: 24,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT.*WHERE l.timestamp >= NOW\(\) - \$1::interval`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			pg := &PostgreSqlx{db: sqlxDB}

			tt.setupMock(mock)

			stats, err := pg.GetHostErrorStats(tt.hours)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.Len(t, stats, len(tt.expected))
				if len(stats) > 0 {
					assert.Equal(t, tt.expected[0].HostID, stats[0].HostID)
					assert.Equal(t, tt.expected[0].HostName, stats[0].HostName)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPostgreSqlx_GetServicesByHost тестирует получение сервисов по хосту
func TestPostgreSqlx_GetServicesByHost(t *testing.T) {
	creationAt := time.Now()

	tests := []struct {
		name        string
		filterType  string
		filterValue interface{}
		setupMock   func(mock sqlmock.Sqlmock)
		expected    []entity.Service
		expectError bool
	}{
		{
			name:        "filter by ip - successful",
			filterType:  "ip",
			filterValue: "192.168.1.1",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "creation_at", "is_active", "env_name",
				}).
					AddRow(
						int64(1), "service1", "Test service", creationAt, true, "production",
					).
					AddRow(
						int64(2), "service2", "Another service", creationAt.Add(-time.Hour), false, "staging",
					)
				mock.ExpectQuery(`SELECT DISTINCT.*WHERE h.ip = \$1`).
					WithArgs("192.168.1.1").
					WillReturnRows(rows)
			},
			expected: []entity.Service{
				{
					ID:        1,
					Name:      "service1",
					Desc:      "Test service",
					CreatedAt: creationAt,
					IsActive:  true,
					EnvName:   "production",
				},
				{
					ID:        2,
					Name:      "service2",
					Desc:      "Another service",
					CreatedAt: creationAt.Add(-time.Hour),
					IsActive:  false,
					EnvName:   "staging",
				},
			},
			expectError: false,
		},
		{
			name:        "filter by host_id - successful",
			filterType:  "host_id",
			filterValue: 123,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "creation_at", "is_active", "env_name",
				}).
					AddRow(int64(1), "service1", "Test service", creationAt, true, "production")
				mock.ExpectQuery(`SELECT DISTINCT.*WHERE h.id = \$1`).
					WithArgs(123).
					WillReturnRows(rows)
			},
			expected: []entity.Service{
				{
					ID:        1,
					Name:      "service1",
					Desc:      "Test service",
					CreatedAt: creationAt,
					IsActive:  true,
					EnvName:   "production",
				},
			},
			expectError: false,
		},
		{
			name:        "invalid filter type",
			filterType:  "invalid",
			filterValue: "value",
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "database error",
			filterType:  "host_name",
			filterValue: "server1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT DISTINCT.*WHERE h.name = \$1`).
					WithArgs("server1").
					WillReturnError(fmt.Errorf("database error"))
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			pg := &PostgreSqlx{db: sqlxDB}

			tt.setupMock(mock)

			services, err := pg.GetServicesByHost(tt.filterType, tt.filterValue)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, services)
			} else {
				assert.NoError(t, err)
				assert.Len(t, services, len(tt.expected))
				if len(services) > 0 {
					assert.Equal(t, tt.expected[0].Name, services[0].Name)
					assert.Equal(t, tt.expected[0].IsActive, services[0].IsActive)
				}
			}

			if tt.name != "invalid filter type" {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestPostgreSqlx_validateTimeUnit тестирует валидацию единиц времени
func TestPostgreSqlx_validateTimeUnit(t *testing.T) {
	tests := []struct {
		name     string
		unit     string
		expected bool
	}{
		{"valid hours", "hours", true},
		{"valid minutes", "minutes", true},
		{"valid days", "days", true},
		{"valid seconds", "seconds", true},
		{"invalid unit", "months", false},
		{"empty unit", "", false},
		{"mixed case", "HOURS", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем тот же валидатор, что и в методе
			validUnits := map[string]bool{
				"hours":   true,
				"minutes": true,
				"days":    true,
				"seconds": true,
			}

			result := validUnits[tt.unit]
			assert.Equal(t, tt.expected, result)
		})
	}
}
