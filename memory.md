# Smart Food Court Management System - Progress Memory & Checkpoint

**Last Updated:** 2026-08-17
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
        8. `foodcourt_staff` (Staff Login Service)
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
*   `staff-login` (Port `8086`)
*   `manager-dashboard` (Port `8087`)
*   `user-dashboard` (Port `8088`)

> [!IMPORTANT]
> **Environment Overrides Fixed**: Updated all 7 microservices' entry points (`main.go`) to use `godotenv.Overload()` instead of `godotenv.Load()`. This isolates service configurations and prevents global system-wide environment variables from hijacking database connections.

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
*   **Direct Authentication Setup**: Removed the OTP verification requirement for staff logins. Staff members (MANAGER, CHEF, ADMIN) now authenticate directly via Bcrypt and receive JWT tokens immediately.
*   **Endpoints:**
    *   `POST /api/staff/login` (Direct password-based authorization whitelisted in Gateway middleware)

---

## 6. Completed Work: Manager Dashboard Service (`manager-dashboard`)
*   **Port:** `8087`
*   **Features:** Financial stats calculations, live active user sessions widget (queried from Redis), inventory management CRUD, and raw stock alerts.

---

## 7. Completed Work: Order & Kitchen Service (`order-kitchen-service`)
*   **Port:** `8084`
*   **Features:** Live order status queue management, chef menu updates, category/dishes management.
*   **Static Asset Serving**: Serves saved files statically at `/api/manager/uploads/` from the local server disk directory `./uploads/`.

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
*   **Razorpay Integration**:
    *   Implemented `POST /api/wallet/recharge/online` endpoint to handle Razorpay payment payloads.
    *   Configured the client wallet page (`src/app/student/wallet/page.tsx`) to pull whitelisted API credentials from `.env.local` and load the overlay checkout overlay window dynamically.
*   **Endpoints:**
    *   `GET /api/wallet/balance` (Query by User ID / Email)
    *   `GET /api/wallet/student` (RFID / search string query)
    *   `GET /api/wallet/history` (Logs of recharge receipts)
    *   `GET /api/wallet/summary` (Stats on daily cashier intakes)
    *   `POST /api/wallet/recharge/nfc`
    *   `POST /api/wallet/recharge/manual`
    *   `POST /api/wallet/recharge/online` (Razorpay)
    *   `POST /api/wallet/deduct` (Safe balance subtraction check)

---

## 10. Completed Work: API Gateway (`api-gateway`)
*   **Port:** `8080`
*   **Features:**
    *   Single-port client interface. Uses dynamic prefix routing (`/api/*any`) to resolve downstream paths without Gin route panics. Whitelists public staff login endpoints from JWT authorization checks.
    *   **Specific Downstream Overlap Checks**: Whitelists and routes `/api/manager/orders`, `/api/manager/menu`, `/api/manager/categories`, and `/api/manager/uploads` to the `order-kitchen-service` (port `8084`) and general metrics/inventory to `manager-dashboard` (port `8087`).
    *   **CORS Support**: Permissive headers configuration to allow client web app integrations.
    *   **JWT Validation**: Checks bearer tokens against a common JWT secret key.
    *   **Blacklist Check**: Queries Redis to deny access to tokens stored during logout.
    *   **Identity Injection**: Pulls user details from Redis cache (`user:id:<userID>`) and sets forwarding headers (`X-User-Id`, `X-User-Role`, `X-User-Name`, `X-User-Email`) before proxying.

---

## 11. Manager Portal Integration Summary
*   ✅ **Dashboard Cards & Overview**: Integrated `src/app/manager/page.tsx` with `/api/manager/overview` to render live today's stats, low-stock alerts, and pending orders. Wiped stale mock pending orders from `local_orders` GORM tables for a clean slate.
*   ✅ **Live Orders Panel**: Connected `src/app/manager/orders/page.tsx` to read dynamic queues (`Preparing in Kitchen`, `Ready to Deliver`, `Dispatched History`) from the database, updated with a 10s auto-refresh interval.
*   ✅ **Menu Management Control**: Integrated the `useFoodStore` Zustand store (`src/stores/food-store.ts`) and dialogs in `src/app/manager/menu/page.tsx` to perform real-time database CRUD operations.
*   ✅ **Base Categories Seeding**: Configured categories seeding (`cat-1` to `cat-8`) to populate category select dropdowns while keeping menu items clean of mock items.
*   ✅ **Base64 Local Image Saver**: Added Base64 string decoding inside `CreateFoodItem` and `UpdateFoodItem` services, writing physical files to `SERVER/order-kitchen-service/uploads/` on disk, returning static server URLs.
*   ✅ **Zustand Storage Quota Fixed**: Disabled local storage `persist` middleware on `useFoodStore` to resolve browser `QuotaExceededError` limits caused by large Base64 images.

---

## 12. Next Steps / Action Items for Next Session
*   **Phase 4: Dining IoT Service (`dining-iot-service` - Port `8083`)**:
    *   Set up GORM tables for table bookings, seat statuses, and occupancy.
    *   Map gateway routing for `/api/dining/*`.
*   **Phase 5: AI Analytics Service (`ai-analytics-service` - Port `8085`)**:
    *   Set up AI menus recommendations and prediction services.
