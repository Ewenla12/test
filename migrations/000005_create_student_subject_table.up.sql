-- +migrate Up
CREATE TABLE student_subject (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students (id) ON DELETE CASCADE,
    subject_id BIGINT NOT NULL REFERENCES subjects (id) ON DELETE CASCADE,
    status application_status NOT NULL DEFAULT 'pending',
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by BIGINT REFERENCES admins (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, subject_id)
);

CREATE INDEX idx_student_subject_status ON student_subject (status);
