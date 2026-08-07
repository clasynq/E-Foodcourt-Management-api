# Smart Food Court Management System - Progress Memory & Checkpoint

**Last Updated:** 2026-08-05
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
        6. `foodcourt_manager` (Manager Dashboard Service)
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

## 5. Completed Work: Staff Login Service (`staff-login`)

### Architecture & Data Models
*   **Model ([staff.go](file:///d:/E-%20Food%20court/SERVER/staff-login/internal/model/staff.go))**:
    *   `StaffMember` GORM table struct.
    *   `InitiateLoginRequest` and `VerifyLoginRequest` (2FA verification flow validation schemas).
*   **Repository ([staff_repository.go](file:///d:/E-%20Food%20court/SERVER/staff-login/internal/repository/staff_repository.go))**:
    *   Data queries (`Create`, `FindByEmail`).
*   **Database Seeding ([config.go](file:///d:/E-%20Food%20court/SERVER/staff-login/internal/config/config.go))**:
    *   Automatically seeds three default demo accounts on startup:
        *   Manager: `rahul@dinesynq.com` (password: `demo123`)
        *   Chef: `ramesh@dinesynq.com` (password: `demo123`)
        *   Admin: `admin@dinesynq.com` (password: `demo123`)

### Business Logic & Redis Integration ([staff_service.go](file:///d:/E-%20Food%20court/SERVER/staff-login/internal/service/staff_service.go))
1.  **InitiateLogin**: Validates credentials, generates secure 6-digit OTP, caches it in Redis (`staff:otp:<email>` with 5 min TTL), and logs/emails it.
2.  **VerifyLogin**: Checks OTP, invalidates/deletes it, and signs JWT claims with `SECRECT_KEY`.
3.  **BlacklistToken**: Invalidates staff JWT tokens inside Redis on logout.

### Handlers & Entrypoint ([staff_handler.go](file:///d:/E-%20Food%20court/SERVER/staff-login/internal/handler/staff_handler.go) / [main.go](file:///d:/E-%20Food%20court/SERVER/staff-login/cmd/server/main.go))
*   Exposed Port: `8086`
*   Registered endpoints:
    *   `POST /api/staff/login/initiate`
    *   `POST /api/staff/login/verify`
    *   `POST /api/staff/logout`

---

## 6. Completed Work: Manager Dashboard Service (`manager-dashboard`)
*   **Port:** `8087`
*   **Database:** `foodcourt_manager`
*   **Architecture & Data Models ([dashboard.go](file:///d:/E-%20Food%20court/SERVER/manager-dashboard/internal/model/dashboard.go))**:
    *   `InventoryItem` (tracks stock status `In Stock`, `Low Stock`, `Out of Stock`).
    *   `LocalOrder` (simulates active and completed orders for dynamic aggregation).
*   **Business Logic & Redis Integration ([dashboard_service.go](file:///d:/E-%20Food%20court/SERVER/manager-dashboard/internal/service/dashboard_service.go))**:
    *   Aggregates today's vs yesterday's metrics (Total Orders, Total Revenue, Avg Prep Time) and calculates percentage differences.
    *   Fetches active user session keys from Redis (`user:email:*`) for real-time active users widget.
    *   Compiles custom multi-ingredient warnings for low stock items.
*   **Handlers & Endpoint Routes ([dashboard_handler.go](file:///d:/E-%20Food%20court/SERVER/manager-dashboard/internal/handler/dashboard_handler.go) / [main.go](file:///d:/E-%20Food%20court/SERVER/manager-dashboard/cmd/server/main.go))**:
    *   Exposes `GET /api/manager/overview` returning stats, pending orders queue, and stock alert models.
    *   *Clean Code Refactoring:* Decoupled Tailwind CSS colors (`bg-blue-500/10` etc.) from API response schema. The backend returns semantic keys (`"orders"`, `"revenue"`, etc.), allowing the client to format styling rules.
    *   Exposes Inventory CRUD endpoints: `GET /api/manager/inventory`, `POST /api/manager/inventory`, `PUT /api/manager/inventory/:id`, `POST /api/manager/inventory/:id/restock`, `DELETE /api/manager/inventory/:id`.

---

## 7. Completed Work: Order & Kitchen Service (`order-kitchen-service`)
*   **Port:** `8084`
*   **Database:** `foodcourt_order`
*   **Architecture & Data Models ([order.go](file:///d:/E-%20Food%20court/SERVER/order-kitchen-service/internal/model/order.go))**:
    *   `Order` GORM schema representing the database source of truth. Includes food item description text (`items`, e.g., `"Biryani, Soda"`), item count, and statuses.
    *   `FoodCategory` and `FoodItem` models mapped for menu management.
*   **Queries & Controllers ([order_repository.go](file:///d:/E-%20Food%20court/SERVER/order-kitchen-service/internal/repository/order_repository.go) / [order_handler.go](file:///d:/E-%20Food%20court/SERVER/order-kitchen-service/internal/handler/order_handler.go))**:
    *   Retrieves active queue orders (status `PENDING`, `PREPARING`, `CONFIRMED`, `READY`), sorted by timestamp.
    *   Updates status of order rows dynamically (so Chef updates instantly reflect in the Manager portal).
*   **Handlers & Endpoint Routes ([main.go](file:///d:/E-%20Food%20court/SERVER/order-kitchen-service/cmd/server/main.go))**:
    *   Exposes Live Order Queue endpoints: `GET /api/manager/orders`, `PUT /api/manager/orders/:id/status`.
    *   Exposes Menu CRUD endpoints: `GET /api/manager/menu`, `GET /api/manager/categories`, `POST /api/manager/menu`, `PUT /api/manager/menu/:id`, `DELETE /api/manager/menu/:id`, `PUT /api/manager/menu/:id/availability`, `PUT /api/manager/menu/:id/stock`.
    *   Fully annotated with Go documentation comments.

---

## 8. Postman Testing Verification
*   ✅ **`POST /api/auth/signup`**: Persists user to PostgreSQL and caches in Redis.
*   ✅ **`POST /api/auth/login`**: Issues JWT token; uses Redis cache first.
*   ✅ **`POST /api/auth/logout`**: Adds current JWT to Redis blacklist.
*   ✅ **`POST /api/auth/send-otp`** / **`verify-otp`**: Sends and validates temporary OTP codes.
*   ✅ **`POST /api/auth/forgot-password`** / **`reset-password`**: Updates user passwords and resets user cache.
*   ✅ **`POST /api/staff/login/initiate`**: Validates credentials and generates staff 2FA code.
*   ✅ **`POST /api/staff/login/verify`**: Validates staff OTP and issues JWT containing roles (`MANAGER`/`CHEF`/`ADMIN`).
*   ✅ **`POST /api/staff/logout`**: Blacklists staff JWT in Redis.
*   ✅ **`GET /api/manager/overview`**: Returns stats, active orders, and stock alerts.
*   ✅ **`GET /api/manager/orders`**: Retrieves live incoming/preparing order queue.
*   ✅ **`PUT /api/manager/orders/:id/status`**: Updates status of a specific order row.
*   ✅ **`GET /api/manager/menu`** / **`categories`**: Retrieves menu dishes and food categories.
*   ✅ **`POST /api/manager/menu`**: Adds a new food item to the menu.
*   ✅ **`PUT /api/manager/menu/:id`**: Updates food item details.
*   ✅ **`DELETE /api/manager/menu/:id`**: Deletes a food item.
*   ✅ **`PUT /api/manager/menu/:id/availability`** / **`stock`**: Toggles availability and updates item stock count.
*   ✅ **`GET /api/manager/inventory`**: Retrieves kitchen raw inventory items.
*   ✅ **`POST /api/manager/inventory`**: Adds a new raw kitchen ingredient.
*   ✅ **`PUT /api/manager/inventory/:id`**: Updates details of an ingredient.
*   ✅ **`POST /api/manager/inventory/:id/restock`**: Restocks ingredient quantity and logs notes.
*   ✅ **`DELETE /api/manager/inventory/:id`**: Deletes an ingredient.

---

## 9. Next Steps / Action Items for Next Session
*   **Phase 2: Add missing fields and profile CRUD to User Service**:
    1.  Add `student_id`, `department`, `avatar`, and `is_active` to the GORM `User` model.
    2.  Implement `GET /api/users/profile`, `PUT /api/users/profile`, `POST /api/users/rfid` (RFID card pairing), and `GET /api/users/by-card/:card_uid` endpoints.
*   **Phase 3: API Gateway (`api-gateway`)**:
    1.  Initialize Go module and Gin framework on port `8080`.
    2.  Implement HTTP reverse proxy to forward client requests on port `8080` to downstream microservice ports (`8081` for User, `8086` for Staff, `8082` for Wallet, etc.).
    3.  Create JWT validation middleware: Verify tokens, check against Redis blacklist, extract `UserID` & `Role`, and inject downstream headers (`X-User-Id`, `X-User-Role`).

