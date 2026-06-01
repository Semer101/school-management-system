# Deploy SMS Frontend on Vercel

## Prerequisites

- Backend already deployed (e.g. Render): `https://school-management-system-70z3.onrender.com`
- GitHub repo connected to Vercel

## Steps

1. **Import project** at [vercel.com/new](https://vercel.com/new) and select this repository.

2. **Set root directory** to `sms-frontend` (not the monorepo root).

3. **Framework preset**: Vite (auto-detected).

4. **Environment variables** (Project → Settings → Environment Variables):

   | Name | Value | Environments |
   |------|--------|--------------|
   | `VITE_API_BASE_URL` | `https://your-backend.onrender.com` | Production, Preview |

   No trailing slash. Example:
   ```
   VITE_API_BASE_URL=https://school-management-system-70z3.onrender.com
   ```

5. **Build settings** (usually auto-filled from `vercel.json`):
   - Build command: `npm run build`
   - Output directory: `dist`

6. **Deploy**. Vercel runs `npm run build`; Vite bakes `VITE_API_BASE_URL` into the bundle.

## Backend CORS

Ensure your Go API allows your Vercel origin. Set on Render (or `.env`):

```
FRONTEND_URL=https://your-app.vercel.app
```

The backend CORS middleware should include this origin for cookies and API calls.

## Cookies / auth

The app uses **HttpOnly cookies** (`sms_access`, `sms_refresh`). The browser must call the API on a host that:

- Shares the same site policy, **or**
- Uses correct `SameSite=None; Secure` on the backend when frontend and API are on different domains.

If login works locally but not on Vercel, update cookie settings in `sms-backend/helpers/auth_cookies.go` for cross-origin production.

## Custom domain

Vercel → Project → Domains → add your domain, then update `FRONTEND_URL` on the backend.

## CLI alternative

```bash
cd sms-frontend
npm i -g vercel
vercel
vercel env add VITE_API_BASE_URL production
vercel --prod
```
