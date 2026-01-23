# Observability System (Логирование и мониторинг)

## 📌 Описание проекта

Проект представляет собой систему для сбора, агрегации и хранения логов от сторонних приложений, написанных на Go. Система состоит из двух основных компонентов:

- **TCP-сервер** для асинхронного приёма логов от клиентов
- **HTTP-сервер** (REST API) для управления системой и получения аналитики


## 🚀 Технологический стек

### Backend
- **Язык**: Go 1.19+ (сервер + клиентская библиотека)
- **Фреймворк**: Echo (REST API) + нативный net (TCP сервер)
- **База данных**: PostgreSQL 14+ с драйвером pgx
- **ORM**: sqlx для работы с БД
- **Аутентификация**: JWT (golang-jwt) + bcrypt
- **Контейнеризация**: Docker + Docker Compose


### Клиентская часть
- TCP-клиент с JSON сериализацией
- Буферизация сообщений через каналы Go
- Поддержка различных уровней логирования
- Автоматическое переподключение

### Инфраструктура
- PostgreSQL в Docker-контейнере
- Конфигурация через .env файлы
- Готовые Docker образы для развертывания



## 🏗️ Архитектура проекта

Структура проекта соответствует стандартным практикам Go:

Для инициализации используется SQL-скрипт и Docker Compose.

## 📦 Основные сущности

| Сущность | Описание | Поля (основные) |
|----------|----------|-----------------|
| **LogEntity** | Основная структура лога | `ID`, `Message`, `Timestamp`, `LevelID`, `ServiceID`, `HostID` |
| **Service** | Сервис-источник логов | `ID`, `Name`, `HostID`, `Description`, `TeamOwner` |
| **Host** | Хост, на котором работает сервис | `ID`, `Name`, `IP`, `CreatedAt` |
| **Environment** | Окружение (dev/stage/prod) | `ID`, `Name`, `Description` |
| **LogLevel** | Уровень логирования | `ID`, `Name`, `Severity`, `ColorCode`, `RetentionDays` |

## 🗄️ База данных

Используется PostgreSQL 14+. Схема базы данных включает следующие таблицы:

```sql
-- Основные таблицы
CREATE TABLE hosts (...);           -- Хосты
CREATE TABLE services (...);        -- Сервисы
CREATE TABLE environments (...);    -- Окружения
CREATE TABLE log_levels (...);      -- Уровни логирования
CREATE TABLE logs (...);            -- Логи
CREATE TABLE users (...);           -- Пользователи API

-- Индексы для производительности
CREATE INDEX idx_logs_timestamp ON logs(timestamp DESC);
CREATE INDEX idx_logs_service_id ON logs(service_id);
CREATE INDEX idx_logs_level_id ON logs(level_id);
```

## 🌐 HTTP-сервер (REST API)

Запускается на порту `:8080`. Поддерживает JWT-аутентификацию.

### Основные эндпоинты:

| Метод | Путь | Описание |
|-------|------|----------|
| POST  | `/register` | Регистрация нового пользователя |
| POST  | `/login` | Авторизация, получение JWT-токена |
| GET   | `/getLogs` | Получение логов (фильтрация по хосту, времени) |
| GET   | `/getErrorStats` | Статистика ошибок по хостам |
| GET   | `/getServices` | Список сервисов на указанном хосте |
| POST  | `/postLogLevel` | Добавление кастомного уровня логирования |
| POST  | `/postEnvironment` | Добавление нового окружения |

## 🔌 TCP-сервер

Принимает логи от клиентов по протоколу TCP на порту `:2000`.

Клиентский логгер (`TCPJSONLogger`) отправляет JSON-сообщения, которые сохраняются в БД.

### Пример использования клиента:

```go
logger, err := NewTCPJSONLogger(
    "localhost:2000",
    "my-service",
    "description",
    "backend-team",
    serviceID,
    environmentID,
    hostIP,
    1000,           
)
```
## 🐳 Запуск проекта

### Клонируйте репозиторий
git clone https://github.com/Kitores/ObservationSystem.git
cd ObservationSystem

### Конфигурационный файл
В configs/ создать файл .env, в котором необходимо указать параметры подключения к БД и JWT-секрет.
```
POSTGRES_DB=db_name
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSL_MODE=disable
JWT_TOKEN="secret-key-change-me"
```
### Запуск сервера с базой данных
```shell
docker-compose up -d
```

### Проверка запущенных контейнеров
```shell
docker-compose ps
```

## Интеграция клиента в ваш Go-проект
### Добавьте зависимость в ваш проект
```shell
go get github.com/Kitores/ObservationSystem
```
Example:
```go
package main

import (
    "github.com/Kitores/ObservationSystem/pkg/tcpConn/tcpClient/logs"
    "time"
)

func main() {
    // Инициализация TCP-логгера
    tcpLogger, err := logs.NewTCPJSONLogger(
        "localhost:2000",          // Адрес сервера ObsSystem
        "user-service",            // Имя вашего сервиса
        "backend-team",            // Название команды
        "192.168.1.100",           // IP адрес хоста
        "Сервис управления пользователями", // Описание
        1000,                      // Размер буфера сообщений
    )

    if err != nil {
        panic("Не удалось подключиться к серверу логов: " + err.Error())
    }
    defer tcpLogger.Close()

    // Пример использования логгера
    tcpLogger.Info("Сервис успешно запущен")
    
    // Логирование с метаданными
    tcpLogger.Debug("Запрос на создание пользователя", 
        map[string]interface{}{
            "user_id": 123,
            "action":  "create",
            "method":  "POST",
            "endpoint": "/api/users",
        })
    
    // Логирование ошибок
    tcpLogger.Error("Не удалось подключиться к базе данных",
        map[string]interface{}{
            "error":    "connection refused",
            "attempts": 3,
            "timeout":  "5s",
        })
    
    // Критические ошибки
    tcpLogger.Fatal("Критическая ошибка: недостаточно памяти")
}
```
