# GoSocial

A production-grade **RESTful social media API** built with Go, following clean architecture and industry best practices. This project demonstrates real-world backend engineering skills including authentication, authorization, caching, rate limiting, database migrations, and more.

---

## Architecture

This project follows a **layered architecture** with clear separation of concerns:

```
cmd/
  api/          # HTTP server, handlers, middleware, routing
  migrate/      # Database migrations & seed scripts
internal/
  auth/         # JWT authentication (generate & validate tokens)
  db/           # Database connection pool setup
  env/          # Environment variable helpers
  mailer/       # Email service (SendGrid integration with templates)
  ratelimiter/  # Fixed-window rate limiter
  store/        # Data access layer (repository pattern)
    cache/      # Redis caching layer
web/            # Static assets / frontend placeholder
```

**Key design decisions:**
- **Repository Pattern** -- each domain (Posts, Users, Comments, Followers, Roles) has its own store interface, making the data layer swappable and testable.
- **Interface-driven** -- `Storage`, `Authenticator`, `Mailer`, and `Limiter` are all interfaces, enabling mock-based testing and loose coupling.
- **Dependency Injection** -- the `application` struct wires all dependencies together in `main.go`, passed down to handlers via closures.
- **Middleware Composition** -- Chi's `middleware.Chain` pattern for auth, rate limiting, logging, CORS, and timeouts.

---

## Tech Stack

| Layer            | Technology                                                     |
| ---------------- | -------------------------------------------------------------- |
| Language         | Go 1.26                                                        |
| Router           | [chi](https://github.com/go-chi/chi)                          |
| Database         | PostgreSQL 17                                                  |
| ORM / Query      | `database/sql` + raw SQL (no ORM)                              |
| DB Migrations    | [golang-migrate](https://github.com/golang-migrate/migrate)   |
| Cache            | Redis 8.8 ([go-redis](https://github.com/redis/go-redis))     |
| Authentication   | JWT (HS256) via [golang-jwt](https://github.com/golang-jwt)   |
| Validation       | [go-playground/validator](https://github.com/go-playground/validator) |
| Logging          | [Zap](https://go.uber.org/zap) (structured, leveled)         |
| Email            | [SendGrid](https://github.com/sendgrid/sendgrid-go)           |
| API Docs         | [Swagger/OpenAPI](https://github.com/swaggo/swag)             |
| Hot Reload       | [Air](https://github.com/air-verse/air)                       |
| Runtime Debug    | `expvar` + [statsviz](https://github.com/arl/statsviz)        |
| Containerization | Docker Compose (PostgreSQL + Redis + Redis Commander)          |
| Testing          | `httptest` + mock stores                                       |

---

## Features

### Authentication & Authorization
- **JWT-based auth** -- register, login, token generation with configurable expiration.
- **Password hashing** -- bcrypt with default cost.
- **Basic Auth** -- for admin/debug endpoints.
- **Role-Based Access Control (RBAC)** -- hierarchical roles (user, moderator, admin) with level-based precedence for post ownership checks.
- **User activation flow** -- invitation tokens with SHA-256 hashing and expiry, email verification via SendGrid.

### Posts (CRUD)
- Create, read, update, delete posts with **optimistic locking** (version field) to prevent lost updates.
- Posts support **tags** (PostgreSQL array type).
- Post retrieval includes **nested comments**.
- Ownership-based authorization -- only the author or a user with sufficient role level can update/delete.

### User Feed
- Paginated feed of posts from followed users + own posts.
- **Filtering** by tags, date range (since/until), and full-text search (ILIKE).
- **Sorting** ascending/descending.
- Returns **comment count** metadata per post.

### Comments
- Create and retrieve comments on posts.
- Comments are nested under posts with user information.

### Follow System
- Follow/unfollow other users.
- Conflict detection on duplicate follows.

### Caching
- **Redis caching layer** for user data with cache-aside pattern.
- Toggle-able via environment variable (`REDIS_ENABLED`).

### Rate Limiting
- **Fixed-window rate limiter** per client IP.
- Configurable requests per time frame.
- Returns `429 Too Many Requests` with `Retry-After` header.

### API Documentation
- Auto-generated **Swagger/OpenAPI** docs via `swag` annotations on handlers.
- Interactive Swagger UI at `/v1/swagger/`.

### Observability
- **Structured logging** with Zap (request ID, method, path, status).
- **`expvar`** -- database stats, goroutine count, version.
- **statsviz** -- real-time runtime visualization at `/v1/debug/statsviz`.

### Reliability
- **Graceful shutdown** -- SIGINT/SIGTERM handling with context timeout.
- **Connection pool management** -- configurable max open/idle connections and idle timeout.
- **Request timeout** -- 60-second global timeout middleware.
- **Request ID** propagation via Chi middleware.
- **CORS** configuration for frontend integration.

---

## REST API Endpoints

All routes are prefixed with `/v1`.

### Health
| Method | Endpoint            | Auth        | Description              |
| ------ | ------------------- | ----------- | ------------------------ |
| GET    | `/health`           | None        | Health check             |

### Authentication
| Method | Endpoint                  | Auth        | Description              |
| ------ | ------------------------- | ----------- | ------------------------ |
| POST   | `/authentication/user`    | None        | Register a new user      |
| POST   | `/authentication/token`   | None        | Login & get JWT token    |

### Users
| Method | Endpoint                      | Auth   | Description              |
| ------ | ----------------------------- | ------ | ------------------------ |
| PUT    | `/users/activate/{token}`    | None   | Activate user account    |
| GET    | `/users/{id}`                | Bearer | Get user profile         |
| PATCH  | `/users/{id}/follow`         | Bearer | Follow a user            |
| PATCH  | `/users/{id}/unfollow`       | Bearer | Unfollow a user          |
| GET    | `/users/{id}/feed`           | Bearer | Get paginated user feed  |

### Posts
| Method | Endpoint        | Auth   | Description              |
| ------ | --------------- | ------ | ------------------------ |
| POST   | `/posts`        | Bearer | Create a post            |
| GET    | `/posts/{id}`   | Bearer | Get post with comments   |
| PATCH  | `/posts/{id}`   | Bearer | Update post (owner/mod)  |
| DELETE | `/posts/{id}`   | Bearer | Delete post (owner/admin)|

### Debug (Basic Auth)
| Method | Endpoint                | Auth   | Description              |
| ------ | ----------------------- | ------ | ------------------------ |
| GET    | `/debug/vars`           | Basic  | Runtime expvar metrics   |
| GET    | `/debug/statsviz`       | Basic  | Runtime visualizations   |

### Docs
| Method | Endpoint            | Auth   | Description              |
| ------ | ------------------- | ------ | ------------------------ |
| GET    | `/swagger/*`        | None   | Swagger UI               |

---

## Database Schema

The project uses **9 sequential migrations**:

1. `init` -- Core tables (users, posts, comments)
2. `add_followers_table` -- Follower relationships
3. `add_indexes` -- Performance indexes
4. `user_invitations` -- Activation token storage
5. `add_activated_to_user` -- User activation status
6. `add_expiry_to_invitations` -- Token expiry
7. `add_time_zone` -- Timestamp timezone support
8. `add_roles_table` -- RBAC roles
9. `alter_users_with_roles` -- Link users to roles

---

## Installation

### Prerequisites
- Go 1.26+
- Docker & Docker Compose
- `migrate` CLI (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)
- `air` for hot reload (`go install github.com/air-verse/air@latest`)
- `swag` for API docs (`go install github.com/swaggo/swag/cmd/swag@latest`)

### Quick Start

1. **Clone the repo:**
   ```bash
   git clone https://github.com/your-username/social.git
   cd social
   ```

2. **Set up environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your values (see Environment Variables below)
   ```

3. **Start infrastructure:**
   ```bash
   make run
   ```
   This starts PostgreSQL + Redis via Docker Compose, then launches the API with hot reload via Air.

4. **Run migrations:**
   ```bash
   make migrate-up
   ```

5. **Seed data (optional):**
   ```bash
   make seed
   ```

6. **Access the API:**
   - API: `http://localhost:8080`
   - Swagger UI: `http://localhost:8080/v1/swagger/`
   - Redis Commander: `http://localhost:8081`

### Manual Steps

```bash
# Start only the database
docker-compose up -d

# Run the server
air

# Run tests
make test

# Generate API docs
make gen-doc

# Create a new migration
make migration <migration_name>

# Rollback migrations
make migrate-down <steps>
```

---

## Environment Variables

| Variable                     | Default                                                    | Description                    |
| ---------------------------- | ---------------------------------------------------------- | ------------------------------ |
| `HTTP_ADDR`                  | `:8080`                                                   | Server listen address          |
| `EXTERNAL_URL`               | `localhost:8080`                                           | Public API URL                 |
| `FRONTEND_URL`               | `http://localhost:3000`                                    | Frontend URL for CORS/email    |
| `ENV`                        | `development`                                              | Environment name               |
| `DB_ADDR`                    | `postgres://admin:adminpassword@localhost/social?sslmode=disable` | PostgreSQL connection string |
| `DB_MAX_OPEN_CONNS`          | `30`                                                       | Max open DB connections        |
| `DB_MAX_IDLE_CONNS`          | `30`                                                       | Max idle DB connections        |
| `DB_MAX_LIFE_TIME`           | `5m`                                                       | Max connection idle time       |
| `REDIS_ADDR`                 | `localhost:6380`                                           | Redis address                  |
| `REDIS_PW`                   | (empty)                                                    | Redis password                 |
| `REDIS_DB`                   | `0`                                                        | Redis database number          |
| `REDIS_ENABLED`              | `false`                                                    | Enable Redis caching           |
| `AUTH_BASIC_USER`            | `admin`                                                    | Basic auth username            |
| `AUTH_BASIC_PASS`            | `admin`                                                    | Basic auth password            |
| `AUTH_TOKEN_SECRET`          | `example`                                                  | JWT signing secret             |
| `FROM_EMAIL`                 | (empty)                                                    | Sender email address           |
| `SENDGRID_API_KEY`           | (empty)                                                    | SendGrid API key               |
| `RATELIMITER_REQUESTS_COUNT` | `20`                                                       | Max requests per window        |
| `RATELIMITER_ENABLE`         | `true`                                                     | Enable rate limiting           |
| `CORS_ALLOWED_ORIGIN`        | `http://*`                                                 | Allowed CORS origins           |

---

## Makefile Commands

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `make run`        | Start infra + server with hot reload     |
| `make test`       | Run all tests                            |
| `make migration`  | Create a new migration file              |
| `make migrate-up` | Run pending migrations                   |
| `make migrate-down` | Rollback migrations                    |
| `make seed`       | Seed the database                        |
| `make gen-doc`    | Generate Swagger documentation           |

---

## Testing

The project uses Go's built-in `testing` package with `httptest` for HTTP testing.

- **Mock stores** -- `store.MockStorage` and `cache.MockCacheStorage` for isolated unit tests.
- **Test authenticator** -- `auth.TestAuthenticator` bypasses JWT for test scenarios.
- **Rate limiter tests** -- verify behavior at the threshold boundary.

```bash
make test
```

---

## Skills Demonstrated

This project showcases the following **backend engineering skills**:

### Core Go
- **Standard library `database/sql`** -- raw SQL queries, connection pooling, transactions (`withTx`), context timeouts.
- **Interface-based design** -- every layer communicates through interfaces for testability and modularity.
- **Goroutines & channels** -- graceful shutdown via signal channels, rate limiter reset goroutines.
- **Context propagation** -- request-scoped values, cancellation, and timeouts throughout the stack.
- **Error handling** -- typed sentinel errors, error wrapping, and structured error responses.

### API Design
- **RESTful conventions** -- proper HTTP methods, status codes, resource naming, and JSON envelope responses.
- **Pagination & filtering** -- query parameter parsing for limit/offset/sort/tags/search/date-range.
- **Content negotiation** -- `Content-Type: application/json` with charset.
- **API documentation** -- Swagger annotations on every handler for auto-generated OpenAPI specs.

### Security
- **JWT authentication** -- HS256 signing, claim validation (exp, iat, nbf, iss, aud), Bearer token extraction.
- **Password hashing** -- bcrypt with configurable cost.
- **RBAC** -- hierarchical role levels for authorization decisions.
- **Request validation** -- struct-tag validation via `go-playground/validator` on all payloads.
- **Input safety** -- `http.MaxBytesReader`, `DisallowUnknownFields` on JSON decoder.
- **CORS** -- configurable origin/method/header allowlisting.

### Database
- **Schema migrations** -- versioned up/down migrations with `golang-migrate`.
- **Optimistic locking** -- version column on posts to detect concurrent update conflicts.
- **PostgreSQL arrays** -- tag storage and `@>` containment queries.
- **Transactions** -- multi-step operations wrapped in DB transactions with rollback.
- **Indexing** -- dedicated migration for performance indexes.

### Infrastructure
- **Docker Compose** -- multi-service setup (PostgreSQL, Redis, Redis Commander) with health checks and persistent volumes.
- **Hot reload** -- Air configuration for development iteration speed.
- **Structured logging** -- Zap with leveled logging, request context, and error details.
- **Rate limiting** -- fixed-window algorithm with per-IP tracking and automatic cleanup.
- **Redis caching** -- cache-aside pattern with fallback to database on miss.

### Engineering Practices
- **Clean architecture** -- clear boundaries between HTTP, business logic, and data access.
- **Configuration management** -- environment variables with sensible defaults via `godotenv`.
- **Graceful shutdown** -- signal handling with cleanup for DB connections, Redis, and HTTP server.
- **Runtime observability** -- `expvar` metrics, `statsviz` visualization, request ID tracking.
- **Testability** -- mock implementations for all external dependencies, HTTP handler testing with `httptest`.

---

## License

MIT
