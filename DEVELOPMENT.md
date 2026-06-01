# Local development — startup order

You do **not** need to re-seed on every run. Use this flow:

## Every day (normal work)

```bash
# Terminal 1 — backend only
cd sms-backend
go run main.go
```

```bash
# Terminal 2 — frontend only
cd sms-frontend
npm run dev
```

- `go run main.go` runs **AutoMigrate** in dev (`ENV` ≠ `production`) and starts the API.
- The database keeps whatever data is already there.

## First setup, or after old bulk seed / slow DB

**Stop the backend**, then run seed once (trim removes legacy bulk rows, then loads ~10 per category):

```bash
cd sms-backend
go run cmd/seed/main.go
```

**Trim only** (delete excess, do not re-seed):

```bash
go run cmd/seed/main.go -trim
```

Then start the backend again:

```bash
go run main.go
```

## Order summary

| Step | Command | When |
|------|---------|------|
| 1 (optional) | `go run cmd/seed/main.go` | First time, or DB still huge / slow |
| 2 | `go run main.go` | Always when developing backend |
| 3 | `npm run dev` | Always when developing frontend |

**Wrong:** running seed while `main.go` is running (two processes writing the same DB).

## Sample logins (after seed)

| Role | Email | Password |
|------|-------|----------|
| Admin | admin@school.et | Admin@1234 |
| Teacher | teacher1@school.et | Teacher@1234 |
| Student | student1@school.et | Student@1234 |
| Parent | parent1@school.et | Parent@1234 |
