# Auth Microservice

HTTP API for user registration, authentication, and profile management. Issues and validates JWTs, manages refresh sessions in Redis, and publishes domain events to Kafka for the rest of the system to react to — it never calls `email` or `broadcasting` directly.

---

## Features

- **Registration & Login**: bcrypt-hashed passwords, JWT access + refresh tokens issued on success.
- **Email verification flow**: a dedicated, purpose-scoped JWT sent via the `email` service; protected routes are gated by `EnsureEmailVerified` until the user confirms.
- **Refresh token rotation**: every `/refresh` call atomically consumes the old token and issues a new one (see [Security Notes](#security-notes)).
- **Logout**: deletes the active refresh session from Redis.
- **User CRUD**: full management of user profiles behind auth middleware.
- **Graceful degradation on partial failure**: if Redis is unreachable during registration, the account and verification email are still real — the request returns an access-only token instead of a hard failure that would leave a "phantom" user (see [Design Decisions](#design-decisions)).
- **Built-in load testing endpoint**: `/api/stress`, intentionally unauthenticated — it's the target of the system's k6 load test and drives the KEDA autoscaling on request rate (see the root README).
- **Prometheus metrics** (`/metrics`) and **health check** (`/api/health`).

---

## Tech Stack

- **Language**: Go 1.25
- **Web framework**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/) (MySQL, PostgreSQL, or SQLite via `DB_DRIVER`)
- **Cache / sessions**: [Redis](https://redis.io/) (`go-redis/v9`)
- **Messaging**: Kafka via [`twmb/franz-go`](https://github.com/twmb/franz-go)
- **Auth**: [`golang-jwt/v5`](https://github.com/golang-jwt/jwt)
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Testing**: [Testify](https://github.com/stretchr/testify) + [Testcontainers](https://testcontainers.com/)

---

## Folder Structure

> For the general architecture patterns used here — the Module Pattern, Repository Pattern, layered `data/handlers/actions/responses` structure, dependency injection via `container`/`app`, typed config, and how to add a new endpoint — see the **[Architecture section of the root README](../../README.md#architecture)**. This section covers only what's specific to `auth`.

```text
internal/
├── bootstrap/           # Wires config, DB, Redis, Kafka publisher; owns graceful startup/shutdown
├── domain/
│   ├── auth/            # Register, Login, Refresh, Logout, VerifyEmail, ResendVerification
│   │   ├── actions/      # One business use case per file
│   │   ├── handlers/     # HTTP handlers (thin: validate → call action → respond)
│   │   └── services/     # JWTService (token issuance/validation)
│   ├── user/             # Profile CRUD (protected)
│   ├── health/            # Liveness
│   └── stress/            # Load-test trigger
├── infrastructure/
│   ├── config/            # Typed, env-driven configuration
│   ├── container/         # DB/Redis/Kafka connections, assembled once at boot
│   ├── database/          # Connection setup per driver
│   ├── redis/              # Session repository (Get/Set/Delete/GetDel)
│   ├── exceptions/          # Environment-aware error responses (hides internals in production)
│   ├── middlewares/          # Auth, EnsureEmailVerified, Prometheus, Logger, Recovery
│   ├── providers/messaging/   # Kafka publisher (reflection-based DTO→topic registry)
│   └── validator/              # Request binding + validation
└── internal/shared/              # go-app-shared submodule (Kafka DTOs, routing keys)
```

---

## API Endpoints

### Auth (`/api/auth`)
| Method | Path | Auth required | Description |
|---|---|---|---|
| POST | `/register` | No | Create account, issue verification email, return tokens |
| POST | `/login` | No | Authenticate, return access + refresh tokens |
| POST | `/refresh` | No (refresh token in body/cookie) | Rotate refresh session, issue new access token |
| POST | `/verify-email` | No (verification token) | Mark account as verified |
| POST | `/resend-verification` | No | Re-send the verification email |
| DELETE | `/logout` | No (refresh token) | Revoke the active refresh session |
| GET | `/validate` | Yes | Validate the current access token |

### Users (`/api/users`) — all require a verified, authenticated user
| Method | Path | Description |
|---|---|---|
| GET | `/` | List users |
| POST | `/` | Create a user manually |
| GET | `/:uuid` | Get a user by UUID |
| PATCH | `/:uuid` | Update a user |
| DELETE | `/:uuid` | Delete a user |

### Other
| Method | Path | Description |
|---|---|---|
| GET | `/api/health` | Liveness check |
| POST | `/api/stress` | Publishes synthetic load to Kafka — public by design, see below |
| GET | `/metrics` | Prometheus metrics |

---

## Security Notes

- **Token purpose scoping**: access tokens and email-verification tokens both use JWT, but they are not interchangeable — each carries a `purpose` claim (`access` vs `email_verification`) that is checked on every validation. A verification token can never be replayed as an access token, and vice versa.
- **Algorithm confusion protection**: `JWTService.ValidateToken` explicitly asserts the signing method is HMAC before trusting the signature, closing the classic `alg: none` bypass.
- **Atomic refresh rotation**: `Refresh` uses Redis `GETDEL` (not `GET` + `DELETE`) to read and invalidate the old session in a single round trip, preventing a race where the same refresh token could be replayed concurrently.
- **Production secret guard**: the app refuses to start in `production` if `AUTH_JWT_SECRET` is empty or left at its insecure default.
- **`/api/stress` is intentionally public.** It exists to drive the k6 load test and the KEDA autoscaling policy on request rate (see the root README's "Load Testing & Autoscaling" section) — it does not touch user data, only publishes synthetic Kafka messages. Do not add auth here; it would defeat the load test's purpose.

---

## Design Decisions

**Graceful degradation over hard failure on `Register`.** If the user row and verification email are already committed but persisting the refresh session to Redis fails, the request does not return a 500. A 500 here would leave the caller with a real, un-loggable-into account (a retry would just hit "email already exists"). Instead, the response degrades to an access-only token (no refresh session) and logs the failure — the user can log in normally once Redis recovers.

**No blacklist of live access tokens.** Logout revokes the *refresh* session in Redis; a short-lived access token issued before logout remains valid until it naturally expires. This is a deliberate trade-off for short access-token TTLs rather than a per-request Redis lookup on every authenticated call — access token lifetime is configured via `AUTH_ACCESS_TOKEN_EXPIRE`.

---

## Messaging — Publishing a New Event

To publish a new event to Kafka without touching any messaging infrastructure code:

**1. Add the DTO** to the shared module (`internal/shared/messaging/kafka/dtos/`):
```go
type PasswordReset struct {
    Email string `json:"email"`
    Token string `json:"token"`
}
```

**2. Register the route** in `setupPublisher` (`internal/bootstrap/api.go`):
```go
publisher.Register(dtos.PasswordReset{}, messaging.Route{
    RoutingKey: constants.RoutePasswordReset,
})
```

**3. Publish from a domain action**, depending only on the `MessagePublisher` interface:
```go
func (a *MyAction) Execute(...) error {
    return a.publisher.Publish(dtos.PasswordReset{Email: email, Token: token})
}
```

The publisher resolves the destination topic by the DTO's Go type via reflection — no switch statements to maintain as event types grow.

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `8080` | HTTP port |
| `DB_DRIVER` | `mysql` | `mysql`, `postgres`, or `sqlite` |
| `AUTH_JWT_SECRET` | — | **Required in production** — rejected if empty/default |
| `AUTH_ACCESS_TOKEN_EXPIRE` | 15 min | Access token TTL |
| `AUTH_REFRESH_TOKEN_EXPIRE` | 7 days | Refresh session TTL |
| `AUTH_EMAIL_VERIFICATION_EXPIRE` | 60 min | Verification token TTL |
| `AUTH_FRONTEND_URL` | — | Base URL used to build the verification link |
| `REDIS_HOST` / `REDIS_PORT` | — | Redis connection |
| `KAFKA_BROKERS` | `kafka:9092` | Kafka bootstrap servers |
| `LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` |

---

## Getting Started

```bash
make init           # copy .env, start containers, migrate, run tests
make run-dev         # run with hot reload (air)
make test             # integration + unit tests via Testcontainers
```

Or from the repo root: `make up`, `make migrate`, `make test` (see the [root README](../../README.md)).
