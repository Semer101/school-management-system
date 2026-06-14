# Deploy SMS Backend on Render

## Prerequisites
- GitHub repo connected to Render
- PostgreSQL database (Render provides this)

## Steps

### 1. Import project
Go to [render.com](https://render.com) and import this repository.

### 2. Settings
- **Root directory**: `sms-backend`
- **Runtime**: Go
- **Build command**: `go build -o main .`
- **Start command**: `./main`

### 3. Environment Variables

| Name | Value |
|------|-------|
| `ENV` | `production` |
| `DATABASE_URL` | *(auto-set from PostgreSQL service)* |
| `JWT_SECRET` | *your-secret-key* |
| `JWT_REFRESH_SECRET` | *your-refresh-secret* |
| `FRONTEND_URL` | `https://your-frontend.vercel.app` |
| `CORS_ORIGINS` | `https://your-frontend.vercel.app` |
| `CSP_CONNECT_SRC` | *(optional, auto-added if FRONTEND_URL set)* |
| `PORT` | `8080` |

### 4. Database
Create a PostgreSQL service in Render and link it to your web service.

### 5. Deploy
Push to GitHub. Render auto-deploys on push.

## CLI Alternative
```bash
# Install Render CLI
npm install -g wrangler
# or use render.yaml with Infrastructure as Code
```