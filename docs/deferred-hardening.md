# Deferred Hardening — Backlog

> Status: **deferred, no functional break** · Captured: 2026-06-03
> Items in this document are correctness/UX/hygiene improvements that were
> intentionally **not** applied in the current iteration. They are tracked here
> so the next person who hits them doesn't have to re-derive the analysis.

## Index

| # | Area                                       | Severity | Owner    | Status |
|---|--------------------------------------------|----------|----------|--------|
| 1 | `axiosClient` token refresh — body read    | Low      | @backend | Open   |
| 2 | `Enrollment.is_retake` flag                | Medium   | @backend | Open   |
| 3 | `GetParents` distinct mismatch             | Medium   | @backend | Open   |
| 4 | Parent archive → re-link workflow          | Low      | TBD      | Ticket |

**Suggested fix order** (smallest, highest leverage first):
`#3` → `#1` → `#2` → `#4` (ticket only).

---

## 1. `axiosClient.ts` token refresh — reads the new token only from the cookie

**File:** `sms-frontend/src/api/axiosClient.ts`
**Server side:** `sms-backend/controllers/auth_controller.go::RefreshToken`
**Server helpers:** `sms-backend/helpers/auth_cookies.go`

### Current state

The 401 interceptor calls the global `axios.post` against
`/api/token/refresh` with `withCredentials: true` and **discards the
response**. The server both:

- sets the new access JWT as the `sms_access` HttpOnly cookie via
  `SetAccessCookie(c, newAccessToken)`, and
- returns `access_token` (and a rotated `refresh_token`) in the JSON body.

The retry of the original request works today only because the browser
auto-attaches the rotated cookie.

### Why it's a deferred-hardening item, not a bug

For pure browser clients the cookie path is fine. The next request
goes through `api`, which already has `withCredentials: true`, so the
new cookie is sent back. Nothing currently breaks.

### Risks (when the cookie path is *not* enough)

- Non-browser API clients (Postman, mobile, server-to-server) do not
  share the browser cookie jar and have no way to get the new token.
- Race condition: the Set-Cookie header and the retried request can
  theoretically be reordered in a hostile proxy setup.
- Future change to refresh-response shape (e.g. dropping the cookie)
  silently breaks the client.

### Recommended fix

Small, backward-compatible. Read the token from the response body and
prefer it when present; also use the `api` instance (not the global
`axios`) for the refresh call so any future baseURL/timeout change
propagates:

```ts
// inside the response interceptor, replacing the `await axios.post(...)` line
try {
  const r = await api.post(refreshUrl, {}, { withCredentials: true })
  // optional: stash on a module-level variable for in-memory access
  // newAccessToken = r?.data?.data?.access_token
  if (r.status !== 200) throw new Error('refresh failed')
  refreshQueue.forEach((cb) => cb())
  refreshQueue = []
  return api(original)
} catch { /* same fallback to /login */ }
```

**Severity:** Low.
**Effort:** ~10 lines, no API changes.
**Rollback:** trivial — keep the cookie path, just add the body read.

---

## 2. `Enrollment` model lacks `is_retake` — retake subjects share codes with first-time

**File:** `sms-backend/models/academic.go`
**Related:** `sms-backend/controllers/academic_ctrl.go` (EnrollStudent, autoEnrollStudentSubjects)
**UI:** `sms-frontend/src/pages/admin/SubjectsPage.tsx`

### Current state

```go
// models/academic.go
type Enrollment struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    StudentID uint      `gorm:"not null;uniqueIndex:idx_enrollment"`
    SubjectID uint      `gorm:"not null;uniqueIndex:idx_enrollment"`
    CreatedAt time.Time `json:"created_at"`
}
```

A student who fails a subject and is re-enrolled next year gets a new
`Enrollment` row pointing at the same `Subject` row. The Subject `code`
(e.g. `MATH-G9`) is shared. UI lists cannot distinguish "first attempt"
from "retake" without parsing timestamps or grades.

### Why the discriminator belongs on the *enrollment*, not the subject

Subject codes legitimately stay stable across cohorts. A retake is a
property of *this student's relationship with this subject in this
year*, not a property of the subject itself.

### Recommended model changes (additive)

```go
// in models/academic.go
type Enrollment struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    StudentID     uint      `gorm:"not null;uniqueIndex:idx_enroll_v2"`
    Student       Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
    SubjectID     uint      `gorm:"not null;uniqueIndex:idx_enroll_v2"`
    Subject       Subject   `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
    IsRetake      bool      `gorm:"default:false"   json:"is_retake"`
    AttemptNumber int       `gorm:"default:1"       json:"attempt_number"` // 1, 2, 3...
    Semester      string    `json:"semester,omitempty"`
    AcademicYear  int       `gorm:"index"           json:"academic_year"`
    RetakeOfID    *uint     `json:"retake_of_id,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
}
```

Replace the existing `idx_enrollment` unique key with a composite
`(student_id, subject_id, academic_year, attempt_number)`. First attempt
and retake in the *same* year must be allowed, so the uniqueness unit
is `(student, subject, year, attempt)` not `(student, subject)`.

### Wiring changes

- `EnrollStudent` / `autoEnrollStudentSubjects` should default
  `IsRetake=false, AttemptNumber=1, AcademicYear=currentYear`.
- Extend `EnrollStudentInput` with
  `{ is_retake, attempt_number, academic_year, retake_of_id }` and/or
  add a separate `POST /api/admin/students/:id/retakes` route
  (admin only).
- Update the duplicate-enrollment check from
  `student_id+subject_id` to
  `student_id+subject_id+academic_year+attempt_number`.
- Update `SubjectsPage` (and any teacher gradebook view) to show a
  small "Retake" badge when `enrollment.is_retake === true` and to
  expose a "Retakes" tab/filter.

### Migration

Add columns nullable with default `false` / `0`. Backfill:

```sql
UPDATE enrollments e
SET academic_year  = s.academic_year,
    attempt_number = 1,
    is_retake      = false
FROM students s
WHERE s.id = e.student_id
  AND e.academic_year IS NULL;
```

Then tighten the unique index. One migration script in
`sms-backend/migration.sql`.

**Severity:** Medium (correctness). Defer if no school is currently
running retakes through the system; otherwise schedule before the next
academic-year rollover.
**Effort:** ~1 day (model + migration + 2 endpoints + UI badge/filter).

---

## 3. `GetParents` count and data queries use different `Distinct` lists

**File:** `sms-backend/controllers/academic_ctrl.go` (function `GetParents`)

### Current state (paraphrased)

```go
// count
dbCount := config.DB.Model(&models.User{}).Where("users.role = ?", models.RoleParent)
// + optional joins on students
dbCount.Distinct("users.id").Count(&total)

// data
db.Distinct("users.id", "users.name", "users.email", "users.phone", "users.is_active", "users.created_at").
    Offset(offset).Limit(limit).Find(&parents)
```

### Why it diverges

- Count `DISTINCT users.id` = number of distinct parents.
- Data `DISTINCT (users.id, users.name, users.email, users.phone, users.is_active, users.created_at)`
  = number of *rows in the cross-product*, which inflates when the
  join to `students` multiplies rows (one parent → many children).
  Different `LIMIT` rows are returned, the reported total is wrong,
  "Page 2 of 27" mismatches reality.
- Worse, `OFFSET`/`LIMIT` on a `DISTINCT` across a join is
  non-deterministic — page boundaries shift between requests.

### Recommended fix — pick the parents once, then preload children

Refactor to the same pattern used elsewhere in this file
(`GetStudents`, `GetTeachers` — single shared query base, *not*
`DISTINCT` on a join):

```go
// 1) Build the same base scoped to role=parent + search.
base := config.DB.Model(&models.User{}).Where("users.role = ?", models.RoleParent)
if q := strings.TrimSpace(c.Query("search")); q != "" {
    // same ILIKE clause on users.name / users.email / s_u.name
}

// 2) One consistent DISTINCT on the parent key, both for count and IDs.
base.Distinct("users.id").Count(&total)

var parentIDs []uint
base.Session(&gorm.Session{}).
    Distinct("users.id").
    Order("users.id DESC").
    Offset(offset).Limit(limit).
    Pluck("users.id", &parentIDs)

// 3) Fetch the user rows by ID — no joins, no distinct, deterministic.
var parents []models.User
config.DB.Where("id IN ?", parentIDs).Order("id DESC").Find(&parents)

// 4) Optionally preload children count:
type childAgg struct{ ParentID uint; Children int64 }
config.DB.Model(&models.Student{}).
    Select("parent_id, COUNT(*) AS children").
    Where("parent_id IN ?", parentIDs).
    Group("parent_id").
    Scan(&childAgg)
// merge into response
```

This makes `total`, `data.length`, and the page boundaries all agree.
Same fix shape as the `GetStudents`/`GetTeachers` "FIX #1/#6" pattern,
just on a `DISTINCT` basis.

**Severity:** Medium (UX — pagination looks broken on the Parents
admin page when a parent has multiple students). Recommend fixing
before the next admin release.
**Effort:** ~30 min. No DB migration.

---

## 4. Parent archive → re-link workflow (audit suggestion, out of scope)

### What's missing today

- `ArchiveParent` soft-deletes the parent user.
- `CreateStudent` accepts only *active* parent IDs.
- There is **no** `Unarchive` endpoint and no UI flow to re-link a
  new student to a previously-archived parent without re-creating the
  parent user from scratch (which creates a duplicate login).

### Recommended flow (target)

1. Admin archives a parent → parent becomes soft-deleted, students
   keep their `parent_id` reference (attendance/finance history
   preserved).
2. Admin tries to create a new student for that parent → blocked at
   the form level with a "Parent is archived" message.
3. Admin opens the parent from the archive list and chooses one of:
   - **Reactivate & link** — restore the same user account.
   - **Link to existing parent** — pick another user (searchable
     dropdown that *includes* archived parents with the same
     "Reactivate & link" affordance).
4. Optionally: an audit log entry for the re-link.

### Recommended next step (NOT to action in this iteration)

- Open a tracking issue titled
  **`Parent archive→re-link workflow`**.
  In the body, capture:
    - today's behavior
    - desired flow (above)
    - user impact and migration implications (existing
      `students.parent_id` references against soft-deleted users)
- Add `// TODO(issue-XXX):` comments in
  `controllers/academic_ctrl.go::ArchiveParent` and
  `CreateStudent` referencing the issue number.
- Re-evaluate at the start of the next quarter — it touches data
  integrity and the parent portal login flow, so it should not be a
  quick fix.

**Severity:** Low (audit deferral).
**Status:** Ticket only. No code change this iteration.

---

## Change log

| Date       | Author | Note                                            |
|------------|--------|-------------------------------------------------|
| 2026-06-03 | @audit | Initial capture; items 1-4 added.               |
