package TcpLogger

import (
	"encoding/json"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"net"
	"sync"
	"time"
)

type TCPJSONLogger struct {
	conn          net.Conn
	mu            sync.Mutex
	address       string
	serviceID     int64
	environmentID int64
	hostIP        string
	reconnect     bool
	buffer        chan []byte
}

func NewTCPJSONLogger(address string, serviceID, environmentID int64, hostIP string, bufferSize int) (*TCPJSONLogger, error) {
	logger := &TCPJSONLogger{
		address:       address,
		serviceID:     serviceID,
		environmentID: environmentID,
		hostIP:        hostIP,
		reconnect:     true,
		buffer:        make(chan []byte, bufferSize),
	}

	// Подключаемся к серверу
	if err := logger.connect(); err != nil {
		return nil, err
	}

	// Запускаем worker для отправки
	go logger.worker()

	return logger, nil
}

func (l *TCPJSONLogger) connect() error {
	conn, err := net.Dial("tcp", l.address)
	if err != nil {
		return fmt.Errorf("failed to connect to log server: %w", err)
	}

	l.mu.Lock()
	l.conn = conn
	l.mu.Unlock()

	return nil
}

func (l *TCPJSONLogger) worker() {
	for data := range l.buffer {
		if err := l.send(data); err != nil {
			// Попытка переподключения
			if l.reconnect {
				l.reconnectWithBackoff()
			}
			// Можно сохранить в локальный файл или кэш
			fmt.Printf("Failed to send log: %v\n", err)
		}
	}
}

func (l *TCPJSONLogger) send(data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.conn == nil {
		return fmt.Errorf("connection is closed")
	}

	// Добавляем разделитель (новую строку)
	data = append(data, '\n')

	// Устанавливаем таймаут
	l.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	_, err := l.conn.Write(data)
	return err
}

func (l *TCPJSONLogger) reconnectWithBackoff() {
	backoff := time.Second
	for i := 0; i < 5; i++ {
		time.Sleep(backoff)
		if err := l.connect(); err == nil {
			return
		}
		backoff *= 2
	}
}

// Методы логирования
func (l *TCPJSONLogger) Info(msg string, metadata ...map[string]interface{}) {
	l.log(1, msg, metadata...)
}

func (l *TCPJSONLogger) Error(msg string, metadata ...map[string]interface{}) {
	l.log(2, msg, metadata...)
}

func (l *TCPJSONLogger) Debug(msg string, metadata ...map[string]interface{}) {
	l.log(3, msg, metadata...)
}

func (l *TCPJSONLogger) log(levelID int64, msg string, metadata ...map[string]interface{}) {
	loggerName := "tcp-json-logger"
	entry := entity.LogEntity{
		Timestamp:     time.Now().UTC(),
		LevelID:       levelID,
		Message:       msg,
		ServiceID:     l.serviceID,
		EnvironmentID: l.environmentID,
		HostIP:        l.hostIP,
		LoggerName:    &loggerName,
	}

	// Добавляем метаданные
	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
	}

	// Сериализуем в JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal log: %v\n", err)
		return
	}

	// Отправляем в буфер для асинхронной отправки
	select {
	case l.buffer <- data:
		// Успешно добавлено в буфер
	default:
		// Буфер переполнен, логируем локально
		fmt.Printf("Log buffer overflow, dropping log: %s\n", msg)
	}
}

func (l *TCPJSONLogger) Close() error {
	l.reconnect = false
	close(l.buffer)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.conn != nil {
		return l.conn.Close()
	}
	return nil
}
