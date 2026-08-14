# Smart Food Court Management System - Progress Memory & Checkpoint

**Last Updated:** 2026-08-14
**Developer:** Pair Programming Session (User + Antigravity)

---

## 1. Project Configuration & Git Status
*   **Git Repository:** Initialized at `d:\E- Food court\SERVER`.
*   **Remote Origin:** `https://github.com/clasynq/E-Foodcourt-Management-api.git`
*   **Root Go Module:** `e-foodcourt-management-api`

---

## 2. Infrastructure Setup
*   **PostgreSQL**:
    *   **Host/Port**: `localhost:5432`
    *   **User/Password**: `postgres` / `suro1234`
    *   **Databases Created & Integrated**:
        1. `foodcourt_auth` (User/Auth Service)
        2. `foodcourt_wallet` (Wallet Service)
        3. `foodcourt_order` (Order/Kitchen Service)
        4. `foodcourt_seating` (Dining IoT Service)
        5. `foodcourt_analytics` (AI Analytics Service)
        6. `foodcourt_manager` (Manager Dashboard Service)
        7. `foodcourt_user_dashboard` (User Dashboard Service)
*   **Redis Cache**:
    *   Running on default port **`6379`**.
    *   Shared across microservices for:
        *   User Session caching (`user:id:<id>` and `user:email:<email>`).
        *   Token Blacklisting (`blacklist:<token>`).

---

## 3. Microservice Environment Configurations
Configured ports, PostgreSQL database connection DSNs, and Redis endpoints for all active microservices:
*   `api-gateway` (Port `8080`)
*   `user-service` (Port `8081`)
*   `wallet-service` (Port `8082`)
*   `dining-iot-service` (Port `8083`)
*   `order-kitchen-service` (Port `8084`)
*   `ai-analytics-service` (Port `8085`)
*   `user-dashboard` (Port `8088`)

---

## 4. Completed Work: User & Authentication Service (`user-service`)
*   **Port:** `8081`
*   **Endpoints:**
    *   `POST /api/auth/signup`
    *   `POST /api/auth/login`
    *   `POST /api/auth/logout` (Blacklists JWT in Redis)
    *   `POST /api/auth/send-otp` / `verify-otp` (2FA/Email validations)
    *   `POST /api/auth/forgot-password` / `reset-password`
*   **Redis Features:** Low-latency read-through caching for login validation, and write-through caching on signup.

---

## 5. Completed Work: Staff Login Service (`staff-login`)
*   **Port:** `8086`
*   **Features:** Multi-role staff 2FA authentication (Manager, Chef, Admin) with Redis OTP validation.

---

## 6. Completed Work: Manager Dashboard Service (`manager-dashboard`)
*   **Port:** `8087`
*   **Features:** Financial stats calculations, live active user sessions widget (queried from Redis), inventory management CRUD, and raw stock alerts.

---

## 7. Completed Work: Order & Kitchen Service (`order-kitchen-service`)
*   **Port:** `8084`
*   **Features:** Live order status queue management, chef menu updates, category/dishes management.

---

## 8. Completed Work: User Dashboard Service (`user-dashboard`)
*   **Port:** `8088`
*   **Database:** `foodcourt_user_dashboard`
*   **Features:**
    *   Aggregates student stats (profile, wallet balance, active orders, reward points).
    *   Queries `wallet-service` over HTTP for live balance.
    *   Queries `order-kitchen-service` for order logs.
    *   Provides fallback values on microservice timeout to guarantee dashboard rendering resilience.
*   **Endpoints:**
    *   `GET /api/student/dashboard/overview`
    *   `POST /api/student/dashboard/rewards` (GORM conflict upserts)

---

## 9. Completed Work: Wallet Service (`wallet-service`)
*   **Port:** `8082`
*   **Database:** `foodcourt_wallet`
*   **Features:**
    *   Manages student digital credits and physical RFID card balances.
    *   Processes cashier manual recharges and card tap NFC recharges.
    *   Handles checkout balance deductions.
*   **Endpoints:**
    *   `GET /api/wallet/balance` (Query by User ID / Email)
    *   `GET /api/wallet/student` (RFID / search string query)
    *   `GET /api/wallet/history` (Logs of recharge receipts)
    *   `GET /api/wallet/summary` (Stats on daily cashier intakes)
    *   `POST /api/wallet/recharge/nfc`
    *   `POST /api/wallet/recharge/manual`
    *   `POST /api/wallet/deduct` (Safe balance subtraction check)

---

## 10. Completed Work: API Gateway (`api-gateway`)
*   **Port:** `8080`
*   **Features:**
    *   Single-port client interface. Uses dynamic prefix routing (`/api/*any`) to resolve downstream paths without Gin route panics.
    *   **CORS Support**: Permissive headers configuration to allow client web app integrations.
    *   **JWT Validation**: Checks bearer tokens against a common JWT secret key.
    *   **Blacklist Check**: Queries Redis to deny access to tokens stored during logout.
    *   **Identity Injection**: Pulls user details from Redis cache (`user:id:<userID>`) and sets forwarding headers (`X-User-Id`, `X-User-Role`, `X-User-Name`, `X-User-Email`) before proxying.

---

## 11. Testing & Verification Summary
*   ✅ **Auth Signup/Login** routed through gateway successfully.
*   ✅ **Token Validations & Blacklist Checks** block invalid tokens or allow legitimate traffic.
*   ✅ **Wallet Balance check & Recharges (NFC/Manual)** successfully update databases and reflect in logs.
*   ✅ **Dashboard Overview aggregation** pulls combined stats from downstream services with full error tolerance.

---

## 12. Next Steps / Action Items for Next Session
*   **Phase 4: Dining IoT Service (`dining-iot-service` - Port `8083`)**:
    *   Set up GORM tables for table bookings, seat statuses, and occupancy.
    *   Map gateway routing for `/api/dining/*`.
*   **Phase 5: AI Analytics Service (`ai-analytics-service` - Port `8085`)**:
    *   Set up AI menus recommendations and prediction services.
*   **Phase 6: Frontend Integration**:
    *   Connect the React/Next.js store endpoints to query port `8080` instead of local mock services.
