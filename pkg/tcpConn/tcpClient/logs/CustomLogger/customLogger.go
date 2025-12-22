package CustomLogger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type JSONLogger struct {
	writer      io.Writer
	serviceID   int64
	environment int64
	hostID      int64
}

type LogEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	ServiceID     int64                  `json:"service_id,omitempty"`
	EnvironmentID int64                  `json:"environment_id,omitempty"`
	HostID        int64                  `json:"host_id,omitempty"`
	LevelID       int64                  `json:"level_id,omitempty"`
	LoggerName    string                 `json:"logger_name,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

func NewJSONLogger(writer io.Writer, serviceID, hostID, environment int64) *JSONLogger {
	return &JSONLogger{
		writer:      writer,
		serviceID:   serviceID,
		environment: environment,
		hostID:      hostID,
	}
}

// Info логирование
func (l *JSONLogger) Info(msg string, fields ...map[string]interface{}) {
	l.log("INFO", 1, msg, nil, fields...)
}

// Error логирование
func (l *JSONLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.log("ERROR", 2, msg, err, fields...)
}

// Debug логирование
func (l *JSONLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.log("DEBUG", 3, msg, nil, fields...)
}

// Warn логирование
func (l *JSONLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.log("WARN", 4, msg, nil, fields...)
}

func (l *JSONLogger) log(level string, levelID int64, msg string, err error, fields ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().UTC(),
		Level:         level,
		Message:       msg,
		ServiceID:     l.serviceID,
		EnvironmentID: l.environment,
		HostID:        l.hostID,
		LevelID:       levelID,
		LoggerName:    "app-logger",
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Добавляем дополнительные поля
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}

	// Сериализуем в JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	// Добавляем новую строку для удобства чтения
	data = append(data, '\n')

	// Записываем
	if _, err := l.writer.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
	}
}

func OpenFile(name string) io.Writer {

	file, err := os.Create(name) // создаем файл
	if err != nil {              // если возникла ошибка
		fmt.Println("Unable to create file:", err)
		//os.Exit(1) // выходим из программы
	}

	file, err = os.OpenFile(name, os.O_WRONLY, 0666)

	if err != nil {
		log.Println(err)
	}

	//defer file.Close()

	return file
}
