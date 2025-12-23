package methods

import (
	"database/sql"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"log"
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
	query := `INSERT INTO logs (service_id, environment_id, host_ip, level_id, 
             message, timestamp, logger_name, recived_at) VALUES (:service_id, :environment_id, :host_ip, :level_id, :message, :timestamp, :logger_name, :recived_at)`
	result, err := pg.db.NamedExec(query, map[string]interface{}{
		"service_id":     logEntity.ServiceID,
		"environment_id": logEntity.EnvironmentID,
		"host_ip":        logEntity.HostIP,
		"level_id":       logEntity.LevelID,
		"message":        logEntity.Message,
		"timestamp":      logEntity.Timestamp,
		"logger_name":    logEntity.LoggerName,
		"received_at":    logEntity.ReceivedAt, //Done! -> Потом можно будет именно время получения по TCP указывать, а не фактическое время сохранения в базу данных
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
func (pg *PostgreSqlx) GetRecentLogsByInterval(hostID int64, value int, unit string) ([]entity.LogEntity, error) {
	validUnits := map[string]bool{
		"hours":   true,
		"minutes": true,
		"days":    true,
		"seconds": true,
	}

	if !validUnits[unit] {
		return nil, fmt.Errorf("invalid time unit: %s", unit)
	}

	query := `
        SELECT 
            l.id,
            l.timestamp,
            l.message,
            l.logger_name,
            l.received_at as received_at,
            l.version,
            
            s.id as service_id,
            s.name as service_name,
            
            h.id as host_id,
            h.name as host_name,
            h.ip as host_ip,
            
            e.id as environment_id,
            e.name as environment_name,
            
            ll.id as level_id,
            ll.name as level_name,
            ll.severity as level_severity,
            ll.color_code as level_color_code
        FROM logs l
        JOIN services s ON l.service_id = s.id
        JOIN log_levels ll ON l.level_id = ll.id
        JOIN hosts h ON l.host_id = h.id
        JOIN environments e ON l.environment_id = e.id
        WHERE l.timestamp >= NOW() - ($1 || ' ' || $2)::interval
          AND ll.name IN ('ERROR', 'FATAL')
          AND h.id = $3  -- Фильтр по конкретному хосту
        ORDER BY l.timestamp DESC
        LIMIT 100
    `

	var logs []entity.LogEntity
	err := pg.db.Select(&logs, query, value, unit, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return logs, nil
}

func (pg *PostgreSqlx) GetHostErrorStats(hours int) ([]entity.HostErrorStats, error) {
	query := `
        SELECT 
            h.id as host_id,
            h.name as host_name,
            COUNT(CASE WHEN ll.name = 'ERROR' THEN 1 END) as error_count,
            COUNT(CASE WHEN ll.name = 'FATAL' THEN 1 END) as fatal_count,
            MAX(l.timestamp) as last_error
        FROM logs l
        JOIN hosts h ON l.host_id = h.id
        JOIN log_levels ll ON l.level_id = ll.id
        WHERE l.timestamp >= NOW() - ($1 || ' hours')::interval
          AND ll.name IN ('ERROR', 'FATAL')
        GROUP BY h.id, h.name
        HAVING COUNT(*) > 0
        ORDER BY error_count + fatal_count DESC
    `

	var stats []entity.HostErrorStats
	err := pg.db.Select(&stats, query, hours)
	return stats, err
}
