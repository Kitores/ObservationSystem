package methods

import (
	"database/sql"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
	"sync"
)

type PostgreSqlx struct {
	db *sqlx.DB
}

var (
	pgInstance *PostgreSqlx
	pgOnce     sync.Once
)

func NewPG(connString string) (*PostgreSqlx, error) {
	pgOnce.Do(func() {
		db, err := sqlx.Connect("pgx", connString)
		if err != nil {
			fmt.Errorf("unable to create conection pool: %w", err)
		}
		pgInstance = &PostgreSqlx{db: db}
		if pgInstance.db == nil {
			log.Fatal("unable to create connection")
		}
		fmt.Println(pgInstance)
	})
	return pgInstance, nil
}

func (pg *PostgreSqlx) Ping() error { return pg.db.Ping() }
func (pg *PostgreSqlx) Close()      {}

// Для TCP сервера
func (pg *PostgreSqlx) SaveLog(logEntity entity.LogEntity) (sql.Result, error) {
	query := `INSERT INTO logs (service_id, environment_id, host_id, level_id, 
             message, timestamp, logger_name, received_at) VALUES (:service_id, :environment_id, :host_id, :level_id, :message, :timestamp, :logger_name, :received_at)`
	result, err := pg.db.NamedExec(query, map[string]interface{}{
		"service_id":     logEntity.ServiceID,
		"environment_id": logEntity.EnvironmentID,
		"host_id":        logEntity.HostID,
		"level_id":       logEntity.LevelID,
		"message":        logEntity.Message,
		"timestamp":      logEntity.Timestamp,
		"logger_name":    logEntity.LoggerName,
		"received_at":    logEntity.ReceivedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to save log: %w", err)
	}
	return result, nil
}

func (pg *PostgreSqlx) RegisterService(newService entity.Service) (int64, int64, error) {
	query := `
        SELECT service_id, host_id 
        FROM register_service($1, $2, $3, $4, $5)
    `

	var result struct {
		ServiceID int64 `db:"service_id"`
		HostID    int64 `db:"host_id"`
	}

	err := pg.db.Get(&result, query,
		newService.Name,
		newService.HostName,
		newService.HostIP,
		*newService.TeamOwner,
		*newService.Desc, // description
	)

	if err != nil {
		return 0, 0, fmt.Errorf("unable to register service: %w", err)
	}

	return result.ServiceID, result.HostID, nil
}

func (pg *PostgreSqlx) UpdateServiceStatus(serviceNameOrHostID interface{}, newStatus bool) error {
	var query string
	switch serviceNameOrHostID.(type) {
	case string: // обновление по имени сервиса
		query = `
            UPDATE services
            SET is_active = $1
            WHERE name = $2`
	case int: // обновление по host_id
		query = `
            UPDATE services
            SET is_active = $1
            FROM hosts
            WHERE hosts.id = services.host_id AND hosts.id = $2`
	default:
		return fmt.Errorf("неподдерживаемый тип аргумента")
	}

	_, err := pg.db.Exec(query, newStatus, serviceNameOrHostID)
	if err != nil {
		return err
	}
	return nil
}

// Для API
// Получение логов за определённый период по заданному фильтру(ip, host_id, host_name)
func (pg *PostgreSqlx) GetRecentLogsByInterval(filterType string, filterValue interface{}, value int, unit string) ([]entity.LogEntity, error) {
	validUnits := map[string]bool{
		"hours":   true,
		"minutes": true,
		"days":    true,
		"seconds": true,
	}

	if !validUnits[unit] {
		return nil, fmt.Errorf("invalid time unit: %s", unit)
	}

	var whereClause string
	var queryParam interface{}

	switch filterType {
	case "ip":
		whereClause = "AND h.ip = $3"
	case "host_id":
		whereClause = "AND h.id = $3"
	case "host_name":
		whereClause = "AND h.name = $3"
	default:
		return nil, fmt.Errorf("invalid filter type: %s", filterType)
	}
	queryParam = filterValue
	query := fmt.Sprintf(`
        SELECT
            l.id,
            l.timestamp,
            l.message,
            l.logger_name,
            l.received_at as received_at,
            l.version,
            
            s.id as service_id,
            
            h.id as host_id,
            h.ip as host_ip,
            
            e.id as environment_id,
            e.name as environment_name,
            
            ll.id as level_id,
            ll.name as level_name,
            ll.severity as level_severity
            
        FROM logs l
        JOIN services s ON l.service_id = s.id
        JOIN log_levels ll ON l.level_id = ll.id
        JOIN hosts h ON l.host_id = h.id
        JOIN environments e ON l.environment_id = e.id
        WHERE l.timestamp >= NOW() - ($1 || ' ' || $2)::interval
            AND ll.name IN ('ERROR', 'FATAL', 'DEBUG')
            %s
        ORDER BY l.timestamp DESC
        LIMIT 100
    `, whereClause)

	// можно менять кароче
	var logs []entity.LogEntity
	err := pg.db.Select(&logs, query, strconv.Itoa(value), unit, queryParam)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return logs, nil
}

// Получение хостов, на которых были ошибки за последний промежуток времени(в часах)
func (pg *PostgreSqlx) GetHostErrorStats(hours int) ([]entity.HostErrorStats, error) {
	intervalStr := fmt.Sprintf("'%d hours'", hours)
	query := `SELECT 
    h.id as host_id,
    h.name as host_name,
    COUNT(CASE WHEN ll.name = 'ERROR' THEN 1 END) as error_count,
    COUNT(CASE WHEN ll.name = 'FATAL' THEN 1 END) as fatal_count,
    MAX(l.timestamp) as last_error
	FROM logs l
	JOIN hosts h ON l.host_id = h.id
	JOIN log_levels ll ON l.level_id = ll.id
	WHERE l.timestamp >= NOW() - $1::interval
	AND ll.name IN ('ERROR', 'FATAL')
	GROUP BY h.id, h.name
	HAVING COUNT(*) > 0
	ORDER BY 
    COUNT(CASE WHEN ll.name = 'ERROR' THEN 1 END) +
    COUNT(CASE WHEN ll.name = 'FATAL' THEN 1 END)
DESC;`

	var stats []entity.HostErrorStats
	err := pg.db.Select(&stats, query, intervalStr)
	if err != nil {
		return nil, err
	}
	return stats, err
}

// Получение сервисов закреплённых за хостом по заданному фильтру(ip, host_id, host_name)
func (pg *PostgreSqlx) GetServicesByHost(filterType string, filterValue interface{}) ([]entity.Service, error) {
	var whereClause string
	var queryParam interface{}

	switch filterType {
	case "ip":
		whereClause = "h.ip = $1"
	case "host_id":
		whereClause = "h.id = $1"
	case "host_name":
		whereClause = "h.name = $1"
	default:
		return nil, fmt.Errorf("invalid filter type: %s", filterType)
	}
	queryParam = filterValue

	query := fmt.Sprintf(`SELECT DISTINCT
	s.id as id,
	s.name as name,
	s.description as description,
	s.creation_at as creation_at,
	s.is_active as is_active,
	
	e.name as env_name
	
	FROM logs l
	JOIN services s ON l.service_id = s.id
	JOIN environments e ON l.environment_id = e.id
	JOIN hosts h ON l.host_id = h.id
	WHERE %s
	ORDER BY s.creation_at DESC
	`, whereClause)

	var services []entity.Service
	err := pg.db.Select(&services, query, queryParam)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	return services, nil
}

// Добавление нового уровня логирования
func (pg *PostgreSqlx) PostNewLogLevel(levelName string, severity int, description string) error {
	query := `INSERT INTO log_levels (name, severity, description) VALUES ($1, $2, $3);`
	_, err := pg.db.Exec(query, levelName, severity, description)
	if err != nil {
		return fmt.Errorf("failed to insert log level: %w", err)
	}
	return nil
}

// Добавление нового окружения
func (pg *PostgreSqlx) PostNewEnvironment(envName string, description string) error {
	query := `INSERT INTO environments (name, description) VALUES ($1, $2);`
	_, err := pg.db.Exec(query, envName, description)
	if err != nil {
		return fmt.Errorf("failed to insert environment: %w", err)
	}
	return nil
}
