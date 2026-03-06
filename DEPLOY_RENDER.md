# 🚀 SIN-ENGINE Deployment Guide

## Quick Deploy Steps

### Step 1: Create PostgreSQL Database on Render
1. Go to https://dashboard.render.com
2. Click **"New +"** → **"PostgreSQL"**
3. Fill in:
   - **Name**: `sin-engine-db`
   - **Database Name**: `sin_engine`
   - **User**: `sin_admin`
4. Click **"Create Database"**
5. Wait for it to provision, then copy the **"Internal Database URL"**

### Step 2: Update API Gateway Service
1. Go to your **sin-engine-api** service
2. Click **"Environment"** tab
3. Add these variables:
   - **DATABASE_URL**: (paste the URL from Step 1)
4. Click **"Create New Deployment"**

### Step 3: Wait for Deploy
- Wait 5-10 minutes for build and deploy
- Check logs for any errors

---

## Alternative: Manual Environment Setup

If auto-deploy doesn't work, add these in Render Dashboard:

```
DATABASE_URL = postgres://sin_admin:password@host:5432/sin_engine
DB_HOST = your-postgres-host.internal
DB_PORT = 5432
DB_USER = sin_admin
DB_PASSWORD = your-password
DB_NAME = sin_engine
DB_SSLMODE = disable
```

---

## Test Your API

After successful deployment:
- **API URL**: `https://sin-engine-api.onrender.com`
- **Health Check**: `https://sin-engine-api.onrender.com/health`

---

## Common Issues

### Issue 1: Build Fails
- Check Dockerfile path is correct: `backend/api-gateway/Dockerfile`
- Make sure Go version is compatible

### Issue 2: Database Connection Failed
- Verify DATABASE_URL is set correctly
- Check PostgreSQL is in same region as web service

### Issue 3: 500 Error
- Check Render logs for error details
- Ensure all environment variables are set
