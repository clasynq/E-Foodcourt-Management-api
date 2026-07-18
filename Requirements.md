# Smart Food Court Management System - API Development Roadmap

This guide outlines the system requirements, library dependencies, and a step-by-step developer roadmap to build the microservices backend for the Smart Food Court Management System.

---

## 1. System Requirements & Prerequisites

To develop, run, and test this microservices cluster locally, you will need the following tools installed on your development machine:

* **Go (Golang):** Version `1.20` or higher (For Gateway, User, Wallet, Order, and Seating services).
* **Python:** Version `3.10` or higher (For the AI Analytics service).
* **PostgreSQL:** Version `14` or higher (Local database instance or running via Docker).
* **Redis:** Version `6` or higher (For fast session storage and recommendation caching).
* **Docker & Docker Compose:** (Optional but highly recommended) To easily run PostgreSQL and Redis servers locally.
* **Postman or Curl:** For API endpoint testing.

---

## 2. Dependencies & Libraries (By Service)

Before starting, initialize packages in each folder with these primary libraries:

### A. Go Services (API Gateway, User, Wallet, Order, Seating)
Run `go mod init <service-name>` in each service folder and install:
* **Web Framework:** `github.com/gin-gonic/gin` or `github.com/gofiber/fiber/v2` (Fast HTTP routing).
* **Database Driver & ORM:** `github.com/jackc/pgx/v5` (Native Postgres driver) or `gorm.io/gorm` and `gorm.io/driver/postgres` (For schema queries).
* **Token Auth:** `github.com/golang-jwt/jwt/v5` (For credentials/JWT generation & verification).
* **Password Hashing:** `golang.org/x/crypto/bcrypt` (For password & pickup PIN hashing).
* **WebSockets:** `github.com/gorilla/websocket` (Required for the API Gateway to route notifications).
* **Caching & Sessions (Redis):** `github.com/redis/go-redis/v9` (For connecting to Redis cache).
* **Environment Variables:** `github.com/joho/godotenv` (For `.env` config files).

### B. AI Analytics Service (Python FastAPI)
Create a Python virtual environment and run `pip install`:
* **FastAPI & Server:** `fastapi`, `uvicorn[standard]`
* **Data Processing & ML:** `pandas`, `numpy`, `scikit-learn`
* **HTTP Client:** `httpx` (To communicate back to Go core services)
* **Caching (Redis client):** `redis` (For retrieving/caching predictions)
* **Environment Configuration:** `python-dotenv`

---

## 3. Step-by-Step API Development Roadmap

Follow this phase-by-phase sequence to build and integrate the microservices:

### Phase 1: Database Setup & Migrations
1. **Database Instances:** 
   * Spin up a PostgreSQL server. 
   * Create 5 separate databases: `auth_db`, `wallet_db`, `order_db`, `seating_db`, and `analytics_db`.
2. **Migrations:** Setup SQL migrations in each Go service directory (or use a migration runner like `golang-migrate` or GORM AutoMigrate). Create tables with native `UUID` types and setup audit fields.
3. **Data Seeding:** Seed default roles (`CUSTOMER`, `VENDOR`, `ADMIN`) in `auth_db` and standard tables in `seating_db`.

### Phase 2: Authentication & User Service (`user-service`)
1. Write routes for User Signup and Login (using Bcrypt password hashing).
2. Generate JWT tokens containing user UUIDs and roles upon successful logins.
3. Write active session endpoints to store token status and blacklisted JWT tokens in Redis.
4. Create NFC registration endpoints to pair a student ID (`card_uid`) to a specific `user_id`.

### Phase 3: API Gateway & WebSocket Router (`api-gateway`)
1. Configure proxy rules using an HTTP reverse proxy in Go. Forward client requests on port `8080` to respective microservice ports.
2. Write JWT validation middleware on the Gateway. Extract the user context and inject it as request headers (`X-User-Id`, `X-User-Role`) downstream.
3. Setup the WebSocket server broker. Save client connection maps against their `user_id` in memory.

### Phase 4: Wallet Service (`wallet-service`)
1. Implement balance inquiry endpoints.
2. Write transaction endpoints:
   * `/api/wallet/debit`: Safe debit checking if `balance >= amount`. 
   * `/api/wallet/credit` (Recharges).
3. Secure the transaction logic using database transactions (SQL `BEGIN ... COMMIT`) to prevent double-spending.

### Phase 5: Seating & IoT Service (`dining-iot-service`)
1. Create CRUD endpoints for dining tables.
2. Implement Table Booking: Set reservation times and update table status to `BOOKED`.
3. Create the IoT sensor webhook handler: `/api/tables/status` (triggered by table sensors).
4. Build a background scheduler task (using standard cron or ticker check) to look for expired reservations (marked `NOSHOW` if vacant for 15 minutes).

### Phase 6: Order & Kitchen Service (`order-kitchen-service`)
1. Implement Vendor and Menu CRUD endpoints (available/unavailable items).
2. Create Order Placement:
   * Customer submits cart. 
   * Order service makes an internal HTTP POST call to Wallet Service to debit the total amount.
   * If wallet payment succeeds, set order to `PAID`, generate a secure 4-digit pickup PIN, and add to the `kitchen_queue`.
3. Build the Kitchen queue dashboard API: Kitchen operators accept, prepare, and update order statuses (Preparing -> Ready).

### Phase 7: AI Analytics Service (`ai-analytics-service`)
1. Write a Python FastAPI server.
2. Implement dummy endpoints representing forecasting logic (e.g., return average queue wait times based on historical queue depth values).
3. Cache recommendation results into Redis for fast retrieval against specific user IDs.

### Phase 8: WebSocket Real-Time Integration
1. Configure Order service to trigger an internal event webhook back to the API Gateway when an order state changes to `READY`.
2. Configure Seating service to notify Gateway when table sensors report occupancy change.
3. The Gateway routes these alerts via WebSockets to keep the frontend client UI in sync in real-time.
