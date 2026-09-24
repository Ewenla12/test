-- +migrate Up
CREATE TYPE student_status AS ENUM ('active', 'graduated', 'suspended', 'dropped_out', 'expelled');
CREATE TYPE application_status AS ENUM ('pending', 'approved', 'rejected');

-- Backs atomic generation of identification numbers like STU-2026-00001
CREATE TABLE id_sequences (
    year INT PRIMARY KEY,
    counter INT NOT NULL DEFAULT 0
);
