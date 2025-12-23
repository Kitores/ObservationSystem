package CustomLogger

import (
	"encoding/json"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"io"
	"log"
	"os"
	"time"
)

type JSONLogger struct {
	writer      io.Writer
	serviceID   int64
	environment int64
	hostIP      string
}

func NewJSONLogger(writer io.Writer, serviceID, environment int64, hostIP string) *JSONLogger {
	return &JSONLogger{
		writer:      writer,
		serviceID:   serviceID,
		environment: environment,
		hostIP:      hostIP,
	}
}

// Info логирование
func (l *JSONLogger) Info(msg string, fields ...map[string]interface{}) {
	l.log(1, msg, nil, fields...)
}

// Error логирование
func (l *JSONLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.log(2, msg, err, fields...)
}

// Debug логирование
func (l *JSONLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(3, msg, nil, fields...)
}

// Warn логирование
func (l *JSONLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(4, msg, nil, fields...)
}

func (l *JSONLogger) log(levelID int64, msg string, err error, metadata ...map[string]interface{}) {
	loggerName := "app-logger"
	entry := entity.LogEntity{
		Timestamp:     time.Now().UTC(),
		Message:       msg,
		ServiceID:     l.serviceID,
		EnvironmentID: l.environment,
		HostIP:        l.hostIP,
		LevelID:       levelID,
		LoggerName:    &loggerName,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Добавляем дополнительные поля
	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
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
