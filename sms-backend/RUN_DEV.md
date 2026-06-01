# Local development — what to run and when

## Every day (normal coding)

**Terminal 1 — backend only:**

```bash
cd sms-backend
go run main.go
```

Leave it running. Auto-migrate runs in dev (`ENV` not set to `production`).

**Terminal 2 — frontend:**

```bash
cd sms-frontend
npm run dev
```

You do **not** need to run the seed every time.

---

## When to run the seed

| Situation | Command |
|-----------|---------|
| First setup or empty DB | `go run cmd/seed/main.go` |
| DB still slow / old 50-student data | `go run cmd/seed/main.go -trim` then `go run cmd/seed/main.go` |
| Only remove bulk data, no re-seed | `go run cmd/seed/main.go -trim` |

**Order:** Database must be reachable (PostgreSQL up). Backend does **not** need to be running for seed.

1. `go run cmd/seed/main.go` — trims excess, then seeds ~10 per category  
2. `go run main.go` — start API  

Or trim once:

```bash
go run cmd/seed/main.go -trim
```

Check logs for counts like `students=10 attendance=...`.

---

## Sample logins (after seed)

| Role | Email | Password |
|------|-------|----------|
| Admin | admin@school.et | Admin@1234 |
| Teacher | teacher1@school.et | Teacher@1234 |
| Student | student1@school.et | Student@1234 |
| Parent | parent1@school.et | Parent@1234 |
