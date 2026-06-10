# SMS Backend — Setup & Run Guide

## Prerequisites
- Go 1.22+
- PostgreSQL 14+
- swag CLI: `go install github.com/swaggo/swag/cmd/swag@latest`

## .env setup
```
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=sms_db
DB_PORT=5432
JWT_SECRET=your-super-secret-key
JWT_REFRESH_SECRET=your-refresh-secret-key
PORT=8080
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your-mailtrap-user
SMTP_PASS=your-mailtrap-pass
SMTP_FROM=noreply@sms.et
```

## Run the server
```bash
# 1. Install dependencies
go mod tidy

# 2. Generate Swagger docs (run every time you add/change annotations)
swag init

# 3. Create database
createdb sms_db

# 4. Start server
go run main.go
```

## Seed the database
```bash
# Trim legacy bulk data, then seed production-quality sample data
go run cmd/seed/main.go

# Trim only (no re-seed)
go run cmd/seed/main.go -trim

# Full sample set (50 students)
go run cmd/seed/main.go -full
```

Creates **3 admins**, **15 teachers**, **25 parents**, and **50 students** with classes, subjects, attendance, grades, notifications, finance, payroll, and locker files.

Re-running the seed is **idempotent**: existing users are matched by email and refreshed (names, phones, student codes, parent links, class assignments). Legacy orphan classes and bulk rows are trimmed first. First run may take several minutes; subsequent runs are faster because attendance and other records are skipped when already present.

Default logins after seeding:

| Role    | Email               | Password     |
|---------|---------------------|--------------|
| Admin   | admin@school.et     | Admin@1234   |
| Admin   | selam@school.et     | Admin@1234   |
| Teacher | teacher1@school.et  | Teacher@1234 |
| Student | student1@school.et  | Student@1234 |
| Parent  | parent1@school.et   | Parent@1234  |

Additional accounts follow the same pattern: `teacher2@school.et`, `student2@school.et`, `parent2@school.et`, etc.

## Install swagger packages
```bash
go get github.com/swaggo/swag
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

## Swagger UI
Visit: http://localhost:8080/swagger/index.html

## API Base URL
http://localhost:8080/api

## Key endpoint groups
| Group       | Base path              | Roles            |
|-------------|------------------------|------------------|
| Auth        | /api/login, /api/register | Public        |
| Profile     | /api/me                | All logged in    |
| Students    | /api/admin/students    | Admin            |
| Teachers    | /api/admin/teachers    | Admin            |
| Parents     | /api/admin/parents     | Admin            |
| Parent portal | /api/parent/children, /api/parent/grades, /api/parent/finance | Parent |
| Attendance  | /api/academics/attendance | Teacher       |
| Grades      | /api/academics/grades  | Teacher          |
| Report Card | /api/academics/reportcard | Student, Parent |
| Locker      | /api/locker            | Student/Teacher  |
| Finance     | /api/finance           | Student/Admin/Parent |
| Broadcast   | /api/admin/notify      | Admin            |
