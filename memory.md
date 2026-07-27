# Smart Food Court Management System - Progress Memory & Checkpoint

**Last Updated:** 2026-07-27
**Developer:** Pair Programming Session (User + Antigravity)

---

## 1. Project Configuration & Git Status
*   **Git Repository:** Initialized at `d:\E- Food court\SERVER`.
*   **Remote Origin:** `https://github.com/clasynq/E-Foodcourt-Management-api.git`
*   **Git Status:** All local changes are committed and pushed to `main`.
*   **Root Go Module:** `e-foodcourt-management-api`

---

## 2. Infrastructure Setup
*   **PostgreSQL**:
    *   **Host/Port**: `localhost:5432`
    *   **User/Password**: `postgres` / `suro1234`
    *   **Databases Created**:
        1. `foodcourt_auth` (User/Auth Service)
        2. `foodcourt_wallet` (Wallet Service)
        3. `foodcourt_order` (Order/Kitchen Service)
        4. `foodcourt_seating` (Dining IoT Service)
        5. `foodcourt_analytics` (AI Analytics Service)
    *   **External Access Configured**:
        *   `listen_addresses = '*'` enabled in `postgresql.conf` (allowing Docker bridge network to connect).
        *   Recommended adding `host all all 0.0.0.0/0 scram-sha-256` in `pg_hba.conf` to allow non-local TCP connections from container subnets.
*   **Redis Cache**:
    *   Running on default port **`6379`**.
    *   Integrated into `user-service` for **Token Blacklisting** and **User Profile Caching**.

---

## 3. Microservice Environment Configurations
Configured `.env` (git-ignored) and `.env.example` templates for all 6 microservices containing database DSNs, ports, and `REDIS_URL`:
*   `api-gateway` (Port `8080`)
*   `user-service` (Port `8081`)
*   `wallet-service` (Port `8082`)
*   `dining-iot-service` (Port `8083`)
*   `order-kitchen-service` (Port `8084`)
*   `ai-analytics-service` (Port `8085`)

---

## 4. Completed Work: User & Authentication Service (`user-service`)

### Architecture & Data Models
*   **Model ([user.go](file:///d:/E-%20Food%20court/SERVER/user-service/internal/model/user.go))**:
    *   `User` table struct with explicit `type:varchar(n)` limits.
    *   `SignupRequest`, `LoginRequest` (DTO validation schemas) and `Claims` (JWT Payload).
    *   `ForgotPasswordRequest` and `ResetPasswordRequest` added for password recovery.
*   **Repository ([user_repository.go](file:///d:/E-%20Food%20court/SERVER/user-service/internal/repository/user_repository.go))**:
    *   Data queries (`Create`, `FindByEmail`, and a newly added `Update` method to save updated user profiles/passwords).

### Business Logic & Redis Integration ([auth_service.go](file:///d:/E-%20Food%20court/SERVER/user-service/internal/service/auth_service.go))
1.  **Bcrypt Hashing**: Secure password hashing and verification during registration, login, and password reset.
2.  **JWT Token Issuance**: Signs tokens containing `UserID` and `Role` with 24-hour expiration.
3.  **Token Blacklisting (Logout)**: Extracts token expiration and stores invalidated tokens in Redis under `blacklist:<token>` with remaining TTL.
4.  **Read-Through / Write-Through Caching**:
    *   On **Signup**: Immediately caches user details in Redis under `user:email:<email>` and `user:id:<id>` for 7 days.
    *   On **Login**: Reads user details from Redis first (*Cache HIT*) to verify password, bypassing PostgreSQL database queries completely! On *Cache MISS*, falls back to PostgreSQL and populates Redis.
5.  **OTP & Password Reset Flows**:
    *   `SendOTP`: Generates and caches secure 6-digit OTPs in Redis, mailing them asynchronously in a goroutine.
    *   `VerifyOTP`: Validates and immediately deletes OTPs from Redis to prevent reuse.
    *   `ForgotPassword`: Reuses the OTP generator to send a reset code.
    *   `ResetPassword`: Verifies the code, updates the password using repository `Update`, and invalidates user cache entries in Redis.

### Handlers & Entrypoint ([auth_handler.go](file:///d:/E-%20Food%20court/SERVER/user-service/internal/handler/auth_handler.go) / [main.go](file:///d:/E-%20Food%20court/SERVER/user-service/cmd/server/main.go))
*   Registered endpoints:
    *   `POST /api/auth/signup`
    *   `POST /api/auth/login`
    *   `POST /api/auth/logout`
    *   `POST /api/auth/send-otp`
    *   `POST /api/auth/verify-otp`
    *   `POST /api/auth/forgot-password`
    *   `POST /api/auth/reset-password`

### Docker Containerization
*   **Dockerfile ([Dockerfile](file:///d:/E-%20Food%20court/SERVER/user-service/Dockerfile))**: Optimized multi-stage Docker build utilizing a `golang:1.26-alpine` builder and `alpine:3.19` runner stage.
*   **Ignore Patterns ([.dockerignore](file:///d:/E-%20Food%20court/SERVER/user-service/.dockerignore))**: Filters out local binaries (`main.exe`), `.env` configuration files, and Git directories.
*   **Container Image**: Successfully built image `dinesynq-user-service:latest` (size ~25MB).

---

## 5. Postman Testing Verification
*   ✅ **`POST /api/auth/signup`**: Persists user to PostgreSQL and caches in Redis.
*   ✅ **`POST /api/auth/login`**: Issues JWT token; uses Redis cache first.
*   ✅ **`POST /api/auth/logout`**: Adds current JWT to Redis blacklist.
*   ✅ **`POST /api/auth/send-otp`** / **`verify-otp`**: Sends and validates temporary OTP codes.
*   ✅ **`POST /api/auth/forgot-password`** / **`reset-password`**: Updates user passwords and resets user cache.

---

## 6. Next Steps / Action Items for Next Session
*   **Phase 2: Add missing fields and profile CRUD to User Service**:
    1.  Add `student_id`, `department`, `avatar`, and `is_active` to the GORM `User` model.
    2.  Implement `GET /api/users/profile`, `PUT /api/users/profile`, `POST /api/users/rfid` (RFID card pairing), and `GET /api/users/by-card/:card_uid` endpoints.
*   **Phase 3: API Gateway (`api-gateway`)**:
    1.  Initialize Go module and Gin framework on port `8080`.
    2.  Implement HTTP reverse proxy to forward client requests on port `8080` to downstream microservice ports (`8081` for User, `8082` for Wallet, etc.).
    3.  Create JWT validation middleware: Verify tokens, check against Redis blacklist, extract `UserID` & `Role`, and inject downstream headers (`X-User-Id`, `X-User-Role`).

