# Student Management API (Go + Fiber + PostgreSQL)

## Setup

1. Install dependencies (requires network access to the Go module proxy —
   not available in the environment this was built in, so this has not
   been run/compiled):
   ```
   go mod tidy
   ```

2. Create a database and set `DATABASE_URL`, e.g.:
   ```
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/school?sslmode=disable"
   export JWT_SECRET="something-long-and-random"
   ```

3. Run migrations. These use golang-migrate's naming convention
   (`NNNNNN_name.up.sql` / `.down.sql`). Install the CLI and run:
   ```
   migrate -database "$DATABASE_URL" -path migrations up
   ```

4. Seed at least one admin manually (no self-registration endpoint exists,
   deliberately — admin accounts shouldn't be self-service):
   ```sql
   INSERT INTO admins (name, email, password_hash)
   VALUES ('First Admin', 'admin@school.test', '<bcrypt hash>');
   ```
   Generate the bcrypt hash with any bcrypt CLI/tool, cost 10+.

5. Run the API:
   ```
   go run ./cmd/api
   ```

## Known gaps / decisions worth revisiting

- **JWT, not sessions.** Tokens can't be revoked server-side before they
  expire (24h). If you need immediate logout/revocation, you'll want a
  token blocklist (e.g. in Redis) — not built here.
- **Reapplying after rejection is blocked.** `student_subject` has a unique
  (student_id, subject_id) constraint, and `Apply` refuses if *any* row
  exists for that pair — including a rejected one. If a rejected student
  should be able to reapply, either allow updating the existing row back
  to `pending`, or drop the uniqueness constraint and track applications
  as a history instead of a single row.
- **No rate limiting on login.** Both login handlers are open to brute
  force as written.
# test
