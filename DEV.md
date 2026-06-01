# Local development

## When to run what

| Situation | Commands |
|-----------|----------|
| **Normal day** (code only) | `cd sms-backend && go run main.go` then `cd sms-frontend && npm run dev` |
| **First setup** or **new DB columns** | Start backend first (AutoMigrate), then seed once |
| **DB still huge / slow** | Run trim (see below) while backend is stopped or running |

You do **not** need to seed every time you start the server.

## Recommended order (first time or after schema change)

```powershell
# Terminal 1 — API (migrates new columns when ENV != production)
cd sms-backend
go run main.go

# Terminal 2 — after DB is up (trim + sample data)
cd sms-backend
go run cmd/seed/main.go
```

## Clean bulk data only (no re-seed)

If the database still has old 50-student / hundreds of attendance rows:

```powershell
cd sms-backend
go run cmd/seed/main.go -trim
```

Then restart the backend if it was running.

## Frontend

```powershell
cd sms-frontend
npm run dev
```

Open http://localhost:5180 — API is proxied to http://localhost:8080.
