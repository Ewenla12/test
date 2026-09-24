-- +migrate Up
CREATE TABLE students (
    id BIGSERIAL PRIMARY KEY,
    identification_number TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    class TEXT,
    status student_status NOT NULL DEFAULT 'active',
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_students_class ON students (class);
CREATE INDEX idx_students_status ON students (status);
CREATE INDEX idx_students_names ON students (first_name, last_name);
