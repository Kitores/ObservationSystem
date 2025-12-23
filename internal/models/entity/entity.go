package entity

import (
	"database/sql"
	"time"
)

type Service struct {
	ID        int64     `db:"id" json:"id,omitempty"`
	Name      string    `db:"name" json:"name"`
	HostName  string    `db:"host_name" json:"host_name"`
	HostID    int64     `db:"host_id" json:"host_id,omitempty"`
	HostIP    string    `db:"host_ip" json:"host_ip"`
	Desc      *string   `db:"desc" json:"description,omitempty"` // nullable
	TeamOwner *string   `db:"team_owner" json:"team_owner,omitempty"`
	CreatedAt time.Time `db:"creation_at" json:"created_at"`
	IsActive  bool      `db:"is_active" json:"is_active"`

	IsFirst bool `json:"is_first"`
}

type Host struct {
	ID   int64   `db:"id" json:"id,omitempty"`
	Name string  `db:"name" json:"name"`
	IP   *string `db:"ip" json:"ip,omitempty"` // nullable
	//Region    *string   `db:"region" json:"region,omitempty"`
	//Zone      *string   `db:"zone" json:"zone,omitempty"`
	//Metadata  *string   `db:"meta_data" json:"metadata,omitempty"` // db: meta_data, json: metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Environment struct {
	ID   int64   `db:"id" json:"id,omitempty"`
	Name string  `db:"name" json:"name"`
	Desc *string `db:"desc" json:"description,omitempty"` // nullable
}

type LogLevel struct {
	ID            int64   `db:"id" json:"id,omitempty"`
	Name          string  `db:"name" json:"name"`
	Severity      int     `db:"severity" json:"severity"`
	ColorCode     *string `db:"color_code" json:"color_code,omitempty"`
	Description   *string `db:"description" json:"description,omitempty"`
	RetentionDays *int    `db:"retention_days" json:"retention_days,omitempty"`
}

type LogEntity struct {
	ID            int64  `db:"id" json:"id,omitempty"`
	ServiceID     int64  `db:"service_id" json:"service_id"`
	EnvironmentID int64  `db:"environment_id" json:"environment_id"`
	HostIP        string `db:"host_ip" json:"host_ip"`
	LevelID       int64  `db:"level_id" json:"level_id"`

	Message    string    `db:"message" json:"message"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"`
	LoggerName *string   `db:"logger_name" json:"logger_name,omitempty"` // nullable

	ReceivedAt time.Time `db:"received_at" json:"received_at,omitempty"`
	Version    int       `db:"version" json:"version,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Error    string                 `json:"error,omitempty"`

	// Для JOIN запросов (опционально, не маппятся напрямую из БД)
	//ServiceName     *string `db:"-" json:"service_name,omitempty"`
	//EnvironmentName *string `db:"-" json:"environment_name,omitempty"`
	//HostName        *string `db:"-" json:"host_name,omitempty"`
	//LevelName       *string `db:"-" json:"level_name,omitempty"`
}

//type LogException struct {
//	ID               int64   `db:"id" json:"id"`
//	LogID            int64   `db:"log_id" json:"log_id"`
//	ExceptionType    *string `db:"exception_type" json:"exception_type,omitempty"`
//	ExceptionMessage *string `db:"exception_message" json:"exception_message,omitempty"`
//}

// service

type LogWithDetails struct {
	ID         int64          `db:"id" json:"id"`
	Timestamp  time.Time      `db:"timestamp" json:"timestamp"`
	Message    string         `db:"message" json:"message"`
	LoggerName sql.NullString `db:"logger_name" json:"logger_name,omitempty"`
	ReceivedAt time.Time      `db:"received_at" json:"received_at"`
	Version    int            `db:"version" json:"version"`

	// Детали сервиса
	ServiceID   int64  `db:"service_id" json:"service_id"`
	ServiceName string `db:"service_name" json:"service_name"`

	// Детали хоста
	HostID   int64          `db:"host_id" json:"host_id"`
	HostName string         `db:"host_name" json:"host_name"`
	HostIP   sql.NullString `db:"host_ip" json:"host_ip,omitempty"`
	//HostRegion sql.NullString `db:"host_region" json:"host_region,omitempty"`
	//HostZone   sql.NullString `db:"host_zone" json:"host_zone,omitempty"`

	// Детали окружения
	EnvironmentID   int64  `db:"environment_id" json:"environment_id"`
	EnvironmentName string `db:"environment_name" json:"environment_name"`

	// Детали уровня
	LevelID       int64          `db:"level_id" json:"level_id"`
	LevelName     string         `db:"level_name" json:"level_name"`
	LevelSeverity int            `db:"level_severity" json:"level_severity"`
	LevelColor    sql.NullString `db:"level_color_code" json:"level_color,omitempty"`
}

type HostErrorStats struct {
	HostID     int64        `db:"host_id" json:"host_id"`
	HostName   string       `db:"host_name" json:"host_name"`
	ErrorCount int          `db:"error_count" json:"error_count"`
	FatalCount int          `db:"fatal_count" json:"fatal_count"`
	LastError  sql.NullTime `db:"last_error" json:"last_error,omitempty"`
}
