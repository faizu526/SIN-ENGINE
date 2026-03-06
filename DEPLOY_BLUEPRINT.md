# 🚀 SIN-ENGINE Deployment Blueprint

## Quick Deploy to Render

### Step 1: Login to Render
```
https://dashboard.render.com
```

### Step 2: Create Web Service
1. Click **"New +"** → **"Web Service"**
2. Connect GitHub: `redmoon/SIN-ENGINE`
3. Configure:

| Setting | Value |
|---------|-------|
| Name | `sin-engine-api` |
| Runtime | **Docker** |
| Dockerfile Path | `backend/api-gateway/Dockerfile` |
| Plan | **Free** |

### Step 3: Environment Variables
Add in **"Advanced"** section:

| Key | Value |
|-----|-------|
| PORT | 8080 |
| GIN_MODE | release |
| LOG_LEVEL | info |

### Step 4: Deploy
Click **"Create Web Service"**

Wait 5-10 minutes for build.

---

## Your Live URL
After deploy: `https://sin-engine-api.onrender.com`

Test: `https://sin-engine-api.onrender.com/health`

Response should be:
```json
{"status": "healthy", "service": "SIN-ENGINE API Gateway"}
```

---

## Troubleshooting

### Build Fails?
- Check Dockerfile path is correct: `backend/api-gateway/Dockerfile`
- Ensure Go modules exist in `backend/api-gateway/`

### 502 Error?
- Wait 2-3 minutes after deploy
- Check logs in Render dashboard

### Need Help?
- Render Docs: https://render.com/docs
- Check logs: Dashboard → Your Service → Logs

---

## Next Steps (After Testing)
1. Add PostgreSQL database from Render dashboard
2. Set JWT_SECRET env var for production
3. Deploy other microservices (auth-service, scanner-service, etc.)

---

## What's Deployed
- **API Gateway** - Main entry point
- **Health Check** - `/health`
- **Auth Routes** - `/api/v1/auth/*`
- **Protected Routes** - Scanner, AI, Recon (require JWT)

---

**Your SIN-ENGINE API Gateway is ready! 🎉**
