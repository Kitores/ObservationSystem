-- Таблица хостов (соответствует entity.Hosts)
CREATE TABLE IF NOT EXISTS hosts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    ip VARCHAR(50) UNIQUE,
--     region VARCHAR(100),
--     zone VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(name, ip)
);

-- Таблица сервисов (соответствует entity.Services)
CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    description TEXT,
    team_owner VARCHAR(100),
    creation_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
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
    ('local', 'Local environment'),
    ('production', 'Production environment'),
    ('staging', 'Staging environment'),
    ('development', 'Development environment')
ON CONFLICT (name) DO NOTHING;



-- Минималистичная версия функции регистрации
CREATE OR REPLACE FUNCTION register_service(
    p_service_name VARCHAR(100),
    p_host_name VARCHAR(255),
    p_host_ip VARCHAR(50),
    p_team_owner VARCHAR(100),
    p_description TEXT DEFAULT NULL
) RETURNS TABLE(service_id INTEGER, host_id INTEGER) AS $$
    -- Вставка хоста (без обработки NULL - будет ошибка если что-то пошло не так)
    WITH host_insert AS (
        INSERT INTO hosts (name, ip, created_at)
        VALUES (p_host_name, p_host_ip, NOW())
        ON CONFLICT (name, ip)
        DO NOTHING
        RETURNING id
    ),
    -- Получаем ID хоста (нового или существующего)
    host_id AS (
        SELECT id FROM host_insert
        UNION ALL
        SELECT id FROM hosts WHERE name = p_host_name AND ip = p_host_ip
        LIMIT 1
    ),
    -- Вставляем сервис
    service_insert AS (
        INSERT INTO services (name, host_id, team_owner, description, is_active, creation_at)
        SELECT p_service_name, id, p_team_owner, p_description, TRUE, NOW()
        FROM host_id
        ON CONFLICT (name)
        DO UPDATE SET
            host_id = EXCLUDED.host_id,
            team_owner = EXCLUDED.team_owner,
            description = EXCLUDED.description,
            is_active = TRUE
        RETURNING id, host_id
    )
SELECT id as service_id, host_id FROM service_insert;
$$ LANGUAGE sql;