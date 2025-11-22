-- Seed данные для тестирования
-- ВАЖНО: Эта миграция предполагает, что таблицы уже созданы
-- и что ID пользователей будут последовательными (1, 2, 3...)

-- Создаем тестовых пользователей
INSERT INTO users (name, is_active) VALUES
    ('Alice', true),
    ('Bob', true),
    ('Charlie', true),
    ('David', true),
    ('Eve', false),  -- неактивный пользователь
    ('Frank', true)
ON CONFLICT DO NOTHING;

-- Создаем команды
INSERT INTO teams (name) VALUES
    ('backend'),
    ('frontend'),
    ('devops')
ON CONFLICT DO NOTHING;

-- Добавляем участников в команды
-- Используем подзапросы для получения ID пользователей по имени
INSERT INTO team_members (team_name, user_id)
SELECT 'backend', id FROM users WHERE name = 'Alice'
ON CONFLICT DO NOTHING;

INSERT INTO team_members (team_name, user_id)
SELECT 'backend', id FROM users WHERE name = 'Bob'
ON CONFLICT DO NOTHING;

INSERT INTO team_members (team_name, user_id)
SELECT 'backend', id FROM users WHERE name = 'Charlie'
ON CONFLICT DO NOTHING;

INSERT INTO team_members (team_name, user_id)
SELECT 'frontend', id FROM users WHERE name = 'David'
ON CONFLICT DO NOTHING;

INSERT INTO team_members (team_name, user_id)
SELECT 'frontend', id FROM users WHERE name = 'Frank'
ON CONFLICT DO NOTHING;

INSERT INTO team_members (team_name, user_id)
SELECT 'devops', id FROM users WHERE name = 'Bob'
ON CONFLICT DO NOTHING;

