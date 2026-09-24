-- +migrate Down
DROP TABLE IF EXISTS id_sequences;
DROP TYPE IF EXISTS application_status;
DROP TYPE IF EXISTS student_status;
