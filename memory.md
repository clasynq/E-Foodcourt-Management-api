# Smart Food Court Management System - Progress Memory & Checkpoint

**Last Updated:** 2026-07-20
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
    *   `User` table struct with explicit `type:varchar(n)` limits (`Name` varchar(100), `Username` varchar(50), `Email` varchar(255), `Phone` varchar(20), `Password` varchar(255), `Role` varchar(20)).
    *   `SignupRequest`, `LoginRequest` (DTO validation schemas) and `Claims` (JWT Payload).
*   **Repository ([user_repository.go](file:///d:/E-%20Food%20court/SERVER/user-service/internal/repository/user_repository.go))**:
    *   Data queries (`Create`, `FindByEmail`).

### Business Logic & Redis Integration ([auth_service.go](file:///d:/E-%20Food%20court/SERVER/user-service/internal/service/auth_service.go))
1.  **Bcrypt Hashing**: Secure password hashing and verification during registration and login.
2.  **JWT Token Issuance**: Signs tokens containing `UserID` and `Role` with 24-hour expiration.
3.  **Token Blacklisting (Logout)**: Extracts token expiration and stores invalidated tokens in Redis under `blacklist:<token>` with remaining TTL.
4.  **Read-Through / Write-Through Caching**:
    *   On **Signup**: Immediately caches user details in Redis under `user:email:<email>` and `user:id:<id>` for 7 days.
    *   On **Login**: Reads user details from Redis first (*Cache HIT*) to verify password, bypassing PostgreSQL database queries completely! On *Cache MISS*, falls back to PostgreSQL and populates Redis.

### Handlers & Entrypoint ([auth_handler.go](file:///d:/E-%20Food%20court/SERVER/user-service/internal/handler/auth_handler.go) / [main.go](file:///d:/E-%20Food%20court/SERVER/user-service/cmd/server/main.go))
*   Registered endpoints:
    *   `POST /api/auth/signup`
    *   `POST /api/auth/login`
    *   `POST /api/auth/logout`

---

## 5. Postman Testing Verification
*   ✅ **`POST /api/auth/signup`**: Tested and returned `201 Created` (Persisted to PostgreSQL and cached in Redis).
*   ✅ **`POST /api/auth/login`**: Tested and returned `200 OK` (Issued signed JWT token; fetched from Redis cache).
*   ✅ **`POST /api/auth/logout`**: Tested with `Authorization: Bearer <token>` header and returned `200 OK` (Blacklisted in Redis).

---

## 6. Next Steps / Action Items for Next Session
*   **Phase 3: API Gateway (`api-gateway`)**:
    1.  Initialize Go module and Gin framework on port `8080`.
    2.  Implement HTTP reverse proxy to forward client requests on port `8080` to downstream microservice ports (`8081` for User, `8082` for Wallet, etc.).
    3.  Create JWT validation middleware: Verify tokens, check against Redis blacklist, extract `UserID` & `Role`, and inject downstream headers (`X-User-Id`, `X-User-Role`).
