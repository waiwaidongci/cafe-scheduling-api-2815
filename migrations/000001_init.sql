CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('manager', 'employee')),
    password_hash TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE shift_types (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT '#64748b',
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_time > start_time)
);

CREATE TABLE business_hours (
    id BIGSERIAL PRIMARY KEY,
    day_of_week SMALLINT NOT NULL UNIQUE CHECK (day_of_week BETWEEN 0 AND 6),
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (close_time > open_time)
);

CREATE TABLE shifts (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    shift_type_id BIGINT NOT NULL REFERENCES shift_types(id) ON DELETE RESTRICT,
    work_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_time > start_time)
);

CREATE INDEX idx_shifts_week_employee ON shifts (work_date, employee_id);
CREATE INDEX idx_shifts_employee_date ON shifts (employee_id, work_date);

INSERT INTO employees (username, name, role, password_hash)
VALUES ('manager', '店长', 'manager', '$2a$10$97yTvvR82GYLKgICIja3QeR/k1JdcMFYsuP.OoxhkznyDbVkbhhj6');

INSERT INTO shift_types (name, color, start_time, end_time)
VALUES
    ('早班', '#0ea5e9', '08:00', '12:00'),
    ('中班', '#64748b', '12:00', '16:00'),
    ('晚班', '#f97316', '16:00', '20:00');

INSERT INTO business_hours (day_of_week, open_time, close_time)
VALUES
    (0, '08:00', '20:00'),
    (1, '08:00', '20:00'),
    (2, '08:00', '20:00'),
    (3, '08:00', '20:00'),
    (4, '08:00', '20:00'),
    (5, '09:00', '18:00'),
    (6, '09:00', '18:00');
