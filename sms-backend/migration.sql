-- ============================================================
--  SMS Backend — Bug-fix migration
--  Run this ONCE on your existing database before deploying
--  the updated backend. Safe to run multiple times (IF NOT EXISTS / DO NOTHING).
-- ============================================================

-- ── 1. refresh_tokens table ───────────────────────────────────────────────────
-- Supports token rotation + revocation (fix for security bug: refresh token
-- was never stored, so it couldn't be validated or invalidated).
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT          NOT NULL,
    expires_at TIMESTAMPTZ   NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id    ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

-- ── 2. Unique constraint on grades ───────────────────────────────────────────
-- FIX #2: Prevents duplicate grade rows from BulkGradeEntry called multiple times.
-- The upsert in the Go code (clause.OnConflict) REQUIRES this constraint to exist.
-- Without it, GORM's OnConflict silently falls back to a plain INSERT, creating duplicates.
-- This migration MUST be run before deploying the updated backend.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_grades_student_subject_type_term_year'
    ) THEN
        ALTER TABLE grades
            ADD CONSTRAINT uq_grades_student_subject_type_term_year
            UNIQUE (student_id, subject_id, type, term, academic_year);
    END IF;
END;
$$;

-- ── 3. Clean up any existing duplicate grades (required before adding the constraint above) ──
-- This deletes older duplicates, keeping the most recently created row.
-- Only runs if there ARE duplicates; safe to run on a clean DB.
DELETE FROM grades
WHERE id NOT IN (
    SELECT MAX(id)
    FROM grades
    GROUP BY student_id, subject_id, type, term, academic_year
);

-- ── 4. Ensure subjects.teacher_id FK is clear about what it references ────────
-- Documents the FK so future devs know TeacherID = teachers.id (NOT users.id).
-- COMMENT ON COLUMN subjects.teacher_id IS 'FK to teachers.id (Teacher table PK, not users.id)';

-- ── 5. Semester-Based Grades Migration ────────────────────────────────────────
-- Rename 'term' column to 'semester' and adjust unique constraints.
-- Replaces 'uq_grades_student_subject_type_term_year' with 'uq_grades_student_subject_type_semester_year'.
DO $$
BEGIN
    -- Rename column term to semester if term exists and semester does not
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'grades' AND column_name = 'term'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'grades' AND column_name = 'semester'
    ) THEN
        ALTER TABLE grades RENAME COLUMN term TO semester;
    END IF;

    -- Drop old unique constraint if exists
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_grades_student_subject_type_term_year'
    ) THEN
        ALTER TABLE grades DROP CONSTRAINT uq_grades_student_subject_type_term_year;
    END IF;

    -- Add new unique constraint if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_grades_student_subject_type_semester_year'
    ) THEN
        ALTER TABLE grades
            ADD CONSTRAINT uq_grades_student_subject_type_semester_year
            UNIQUE (student_id, subject_id, type, semester, academic_year);
    END IF;
END;
$$;

-- Migrate values in semester column
UPDATE grades SET semester = 'Semester 1' WHERE semester IN ('Term1', 'Term 1');
UPDATE grades SET semester = 'Semester 2' WHERE semester IN ('Term2', 'Term 2');
UPDATE grades SET semester = 'Semester 3' WHERE semester IN ('Term3', 'Term 3');

-- ── Done ──────────────────────────────────────────────────────────────────────
-- After running this migration, restart the backend so GORM AutoMigrate
-- picks up the new RefreshToken model and creates any missing columns.

