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
go run cmd/seed/main.go
```

Default logins after seeding:
- Admin:   admin@sms.et / admin123
- Teacher: abebe.g@sms.et / teacher123
- Student: student1@sms.et / student123

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
| Attendance  | /api/academics/attendance | Teacher       |
| Grades      | /api/academics/grades  | Teacher          |
| Report Card | /api/academics/reportcard | All           |
| Locker      | /api/locker            | Student/Teacher  |
| Finance     | /api/finance           | Student/Admin    |
| Broadcast   | /api/admin/notify      | Admin            |
