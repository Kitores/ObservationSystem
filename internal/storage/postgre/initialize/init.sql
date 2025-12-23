-- Таблица сервисов (соответствует entity.Services)
CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    host_ip VARCHAR(50) NOT NULL REFERENCES hosts(ip) ON DELETE CASCADE
    description TEXT,
    team_owner VARCHAR(100),
    creation_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Таблица хостов (соответствует entity.Hosts)
CREATE TABLE IF NOT EXISTS hosts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    ip VARCHAR(50),
--     region VARCHAR(100),
--     zone VARCHAR(100),
    meta_data TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(name, ip)
);

-- Таблица окружений (соответствует entity.Environments)
CREATE TABLE IF NOT EXISTS environments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description  TEXT
);

-- Таблица уровней логирования (соответствует entity.LogLevels)
CREATE TABLE IF NOT EXISTS log_levels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE,
    severity INTEGER NOT NULL,       -- для сортировки (10=DEBUG, 20=INFO и т.д.)
    color_code VARCHAR(7),           -- для UI
    description TEXT,
    retention_days INTEGER           -- политика хранения для разных уровней
);

-- Основная таблица логов (соответствует entity.Logs)
CREATE TABLE IF NOT EXISTS logs (
    id BIGSERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    level_id INTEGER NOT NULL REFERENCES log_levels(id) ON DELETE CASCADE,

    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    logger_name VARCHAR(255),

    received_at TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE,
    password VARCHAR(20) NOT NULL
);
-- -- Таблица исключений (соответствует entity.LogException)
-- CREATE TABLE IF NOT EXISTS log_exceptions (
--     id BIGSERIAL PRIMARY KEY,
--     log_id BIGINT NOT NULL REFERENCES logs(id) ON DELETE CASCADE,
--     exception_type VARCHAR(255),
--     exception_message TEXT
-- );

-- -- Создание индексов для производительности
-- CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
-- CREATE INDEX IF NOT EXISTS idx_logs_service_id ON logs(service_id);
-- CREATE INDEX IF NOT EXISTS idx_logs_environment_id ON logs(environment_id);
-- CREATE INDEX IF NOT EXISTS idx_logs_host_id ON logs(host_id);
-- CREATE INDEX IF NOT EXISTS idx_logs_level_id ON logs(level_id);
-- CREATE INDEX IF NOT EXISTS idx_logs_recived_at ON logs(recived_at);
--
-- CREATE INDEX IF NOT EXISTS idx_log_exceptions_log_id ON log_exceptions(log_id);

-- Вставка начальных данных для уровней логирования
INSERT INTO log_levels (name, severity) VALUES
    ('DEBUG', 10),
    ('INFO', 20),
    ('WARN', 30),
    ('ERROR', 40),
    ('FATAL', 50)
ON CONFLICT (name) DO NOTHING;

-- Вставка начальных данных для окружений
INSERT INTO environments (name, description) VALUES
    ('local', 'Local environment')
    ('production', 'Production environment'),
    ('staging', 'Staging environment'),
    ('development', 'Development environment')
ON CONFLICT (name) DO NOTHING;


