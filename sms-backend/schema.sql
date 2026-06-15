-- ============================================================
--  SMS Backend — Complete Database Schema
--  Generated from GORM models (GORM v1.25 / pgx v5).
--
--  Run this ONCE in the Supabase SQL Editor on a fresh project.
--  Every statement is idempotent (IF NOT EXISTS) so it is safe
--  to re-run without dropping existing data.
--
--  Table creation order respects foreign-key dependencies:
--    users → teachers → classes / subjects → students →
--    enrollments / attendances / grades / locker_files /
--    transactions / payrolls / notifications →
--    notification_receipts / refresh_tokens / password_reset_otps
-- ============================================================


-- ── 1. users ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL    PRIMARY KEY,
    name       TEXT         NOT NULL,
    email      TEXT         NOT NULL,
    password   TEXT         NOT NULL,
    role       TEXT         NOT NULL DEFAULT 'Student',
    phone      TEXT,
    avatar_url TEXT,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
-- Partial unique index: allows re-use of email after soft-delete
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email      ON users (email) WHERE deleted_at IS NULL;
CREATE        INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);


-- ── 2. teachers ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS teachers (
    id            BIGSERIAL    PRIMARY KEY,
    user_id       BIGINT       NOT NULL,
    teacher_code  TEXT         NOT NULL,
    qualification TEXT,
    department    TEXT,
    joined_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT fk_teachers_user FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_user_id      ON teachers (user_id)      WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_teacher_code ON teachers (teacher_code) WHERE deleted_at IS NULL;
CREATE        INDEX IF NOT EXISTS idx_teachers_deleted_at   ON teachers (deleted_at);


-- ── 3. classes ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS classes (
    id          BIGSERIAL    PRIMARY KEY,
    name        TEXT         NOT NULL,
    grade_level BIGINT       NOT NULL DEFAULT 9,
    section     TEXT         NOT NULL DEFAULT 'A',
    stream      TEXT,
    status      TEXT         NOT NULL DEFAULT 'Active',
    year        BIGINT       NOT NULL,
    teacher_id  BIGINT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT fk_classes_teacher FOREIGN KEY (teacher_id) REFERENCES teachers (id)
);
CREATE INDEX IF NOT EXISTS idx_classes_deleted_at ON classes (deleted_at);


-- ── 4. subjects ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS subjects (
    id          BIGSERIAL    PRIMARY KEY,
    name        TEXT         NOT NULL,
    code        TEXT         NOT NULL,
    grade_level BIGINT       NOT NULL DEFAULT 0,
    stream      TEXT,
    status      TEXT         NOT NULL DEFAULT 'Active',
    teacher_id  BIGINT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT fk_subjects_teacher FOREIGN KEY (teacher_id) REFERENCES teachers (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subjects_code       ON subjects (code)       WHERE deleted_at IS NULL;
CREATE        INDEX IF NOT EXISTS idx_subjects_deleted_at ON subjects (deleted_at);


-- ── 5. students ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS students (
    id               BIGSERIAL    PRIMARY KEY,
    user_id          BIGINT       NOT NULL,
    parent_id        BIGINT       NOT NULL DEFAULT 0,
    class_id         BIGINT,
    student_code     TEXT         NOT NULL,
    parent_name      TEXT,
    parent_email     TEXT,
    parent_phone     TEXT,
    date_of_birth    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    enrolled_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    stream           TEXT         NOT NULL DEFAULT '',
    grade_level      BIGINT       NOT NULL DEFAULT 9,
    promotion_status TEXT         NOT NULL DEFAULT 'normal',
    academic_year    BIGINT       NOT NULL DEFAULT 2025,
    deleted_at       TIMESTAMPTZ,
    CONSTRAINT fk_students_user  FOREIGN KEY (user_id)  REFERENCES users    (id),
    CONSTRAINT fk_students_class FOREIGN KEY (class_id) REFERENCES classes  (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_user_id      ON students (user_id)      WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_student_code ON students (student_code) WHERE deleted_at IS NULL;
CREATE        INDEX IF NOT EXISTS idx_students_deleted_at   ON students (deleted_at);


-- ── 6. enrollments ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS enrollments (
    id         BIGSERIAL    PRIMARY KEY,
    student_id BIGINT       NOT NULL,
    subject_id BIGINT       NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_enrollments_student FOREIGN KEY (student_id) REFERENCES students (id),
    CONSTRAINT fk_enrollments_subject FOREIGN KEY (subject_id) REFERENCES subjects (id),
    CONSTRAINT idx_enrollment          UNIQUE (student_id, subject_id)
);


-- ── 7. attendances ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS attendances (
    id         BIGSERIAL    PRIMARY KEY,
    student_id BIGINT       NOT NULL,
    subject_id BIGINT,
    date       TIMESTAMPTZ  NOT NULL,
    status     TEXT         NOT NULL,
    notes      TEXT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_attendances_student FOREIGN KEY (student_id) REFERENCES students (id),
    CONSTRAINT fk_attendances_subject FOREIGN KEY (subject_id) REFERENCES subjects (id)
);


-- ── 8. grades ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS grades (
    id            BIGSERIAL    PRIMARY KEY,
    student_id    BIGINT       NOT NULL,
    subject_id    BIGINT       NOT NULL,
    teacher_id    BIGINT       NOT NULL,
    score         NUMERIC      NOT NULL,
    max_score     NUMERIC      NOT NULL DEFAULT 100,
    type          TEXT         NOT NULL,
    semester      TEXT         NOT NULL,
    academic_year BIGINT       NOT NULL,
    remarks       TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_grades_student FOREIGN KEY (student_id) REFERENCES students (id),
    CONSTRAINT fk_grades_subject FOREIGN KEY (subject_id) REFERENCES subjects (id),
    CONSTRAINT uq_grades_student_subject_type_semester_year
        UNIQUE (student_id, subject_id, type, semester, academic_year)
);


-- ── 9. locker_files ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS locker_files (
    id          BIGSERIAL    PRIMARY KEY,
    student_id  BIGINT       NOT NULL,
    file_name   TEXT         NOT NULL,
    file_path   TEXT         NOT NULL,
    file_size   BIGINT       NOT NULL DEFAULT 0,
    file_type   TEXT,
    category    TEXT,
    is_public   BOOLEAN      NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_locker_files_student FOREIGN KEY (student_id) REFERENCES students (id)
);


-- ── 10. transactions ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS transactions (
    id                BIGSERIAL    PRIMARY KEY,
    student_id        BIGINT       NOT NULL,
    parent_id         BIGINT,
    amount            NUMERIC      NOT NULL,
    receipt_id        TEXT         NOT NULL,
    receipt_image_url TEXT,
    type              TEXT         NOT NULL,
    status            TEXT         NOT NULL DEFAULT 'Pending',
    description       TEXT,
    rejection_notes   TEXT,
    created_by        BIGINT       NOT NULL DEFAULT 0,
    verified_by       BIGINT       NOT NULL DEFAULT 0,
    verified_at       TIMESTAMPTZ,
    academic_year     BIGINT       NOT NULL DEFAULT 0,
    semester          TEXT,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_transactions_student FOREIGN KEY (student_id) REFERENCES students (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_receipt_id ON transactions (receipt_id);


-- ── 11. payrolls ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payrolls (
    id         BIGSERIAL    PRIMARY KEY,
    teacher_id BIGINT       NOT NULL,
    amount     NUMERIC      NOT NULL,
    month      BIGINT       NOT NULL,
    year       BIGINT       NOT NULL,
    status     TEXT         NOT NULL DEFAULT 'Pending',
    paid_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_payrolls_teacher FOREIGN KEY (teacher_id) REFERENCES teachers (id)
);


-- ── 12. notifications ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS notifications (
    id           BIGSERIAL    PRIMARY KEY,
    title        TEXT         NOT NULL,
    body         TEXT         NOT NULL,
    target_roles TEXT         NOT NULL DEFAULT '',
    sender_id    BIGINT       NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_notifications_sender FOREIGN KEY (sender_id) REFERENCES users (id)
);


-- ── 13. notification_receipts ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS notification_receipts (
    id              BIGSERIAL    PRIMARY KEY,
    user_id         BIGINT       NOT NULL,
    notification_id BIGINT       NOT NULL,
    is_read         BOOLEAN      NOT NULL DEFAULT FALSE,
    read_at         TIMESTAMPTZ,
    CONSTRAINT fk_nr_user         FOREIGN KEY (user_id)         REFERENCES users         (id),
    CONSTRAINT fk_nr_notification FOREIGN KEY (notification_id) REFERENCES notifications (id)
);
CREATE INDEX IF NOT EXISTS idx_notification_receipts_user_id         ON notification_receipts (user_id);
CREATE INDEX IF NOT EXISTS idx_notification_receipts_notification_id ON notification_receipts (notification_id);


-- ── 14. refresh_tokens ────────────────────────────────────────────────────────
-- Non-unique: one user can have multiple sessions (multiple browser tabs/devices).
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL,
    token_hash TEXT         NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id    ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);


-- ── 15. password_reset_otps ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS password_reset_otps (
    id         BIGSERIAL    PRIMARY KEY,
    email      TEXT         NOT NULL,
    otp_hash   TEXT         NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    used       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_password_reset_otps_email ON password_reset_otps (email);


-- ── Done ──────────────────────────────────────────────────────────────────────
-- After running this, restart the Render backend.
-- GORM AutoMigrate will run on startup, add any missing columns, and
-- EnsureDefaultUsers will seed the four default login accounts.
