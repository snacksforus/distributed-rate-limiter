# Distributed Rate Limiter

This project is a demonstration of a distributed rate limiter implemented in Go, using 
Redis as a 
shared state backend to coordinate rate limiting across multiple application instances.

## Architecture

```
             HTTP Request
                  │
                  ▼
┌─────────────────────────────────────┐
│       HTTP Server (DRL_PORT)        │
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│         Rate Limit Middleware       │
│   (extracts client IP, calls limiter│
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│           Rate Limiter              │
│  (fixed window algorithm, fail-open)│
└─────────────────┬───────────────────┘
                  │
                  ▼ 
┌─────────────────────────────────────┐
│         Storage (Redis)             │
│  (atomic incr + TTL via Lua script) │
└─────────────────────────────────────┘
```

### Key Components

- API: Demo HTTP endpoint that return a JSON success message.
- Middleware: Sits in front of all routes.  Extracts client IP, invokes the rate 
  limiter, and returns a 429 error if the limit is exceeded.
- Rate Limiter: Implements a fixed window rate limit algorithm.  Fails open if an 
  error is encountered.
- Storage: Abstraction layer for Redis.  Uses a Lua script to atomically increment 
  client request counters.
- Config: Application configuration using environment variables.

### Fixed Window Algorithm

The rate limiter uses a fixed window algorithm.  Each request increments a Redis 
counter using the client's IP address.  The first request sets the TTL to 
`DRL_WINDOW_SIZE_SEC`.  When the counter exceeds `DRL_RATE_LIMIT` the request is 
rejected.  The counter is reset when the Redis key expires.

Note: The fixed window algorithm allows burst of requests at window boundaries.  This 
is sufficient for demonstrations purposes, but a sliding window algorithm would 
prevent the bursts of requests.

## Configuration

All configuration is provided via environment variables.

| Variable             | Default     | Description                          |
|----------------------|-------------|--------------------------------------|
| `DRL_HOSTNAME`       | empty       | API server bind hostname             |
| `DRL_PORT`           | `8080`      | API server port (1–65535)            |
| `DRL_RATE_LIMIT`     | `10`        | Max requests allowed per window      |
| `DRL_WINDOW_SIZE_SEC`| `10`        | Window duration in seconds           |
| `DRL_REDIS_HOSTNAME` | `drl-redis` | Redis hostname (required)            |
| `DRL_REDIS_PORT`     | `6379`      | Redis port (1–65535)                 |
| `DRL_REDIS_PASSWORD` | empty       | Redis password (optional)            |

## Running the Application

### Prerequisites

- Docker
- Docker Compose
- Make

### Start

```bash
make build-network   # create the drl-network Docker network (first time only)
make run             # start Redis and the API server
```

The API server listens on port `8080` by default.

### Stop

```bash
make stop
```

### API

```bash
curl http://localhost:8080/api
```

**200 OK — request allowed:**
```json
{"success":true,"error":null}
```

**429 Too Many Requests — rate limit exceeded:**
```json
{"success":false,"error":{"code":"TOO_MANY_REQUESTS","message":"Exceeded request rate limit"}}
```

## Running the Tests

Tests run inside Docker and require a live Redis instance, which Docker Compose spins up automatically.

```bash
make test
```
