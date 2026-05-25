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

| Variable                     | Default     | Description                        |
|------------------------------|-------------|------------------------------------|
| `DRL_HOSTNAME`               | empty       | API server bind hostname           |
| `DRL_PORT`                   | `8080`      | API server port (1–65535)          |
| `DRL_RATE_LIMIT`             | `10`        | Max requests allowed per window    |
| `DRL_WINDOW_SIZE_SEC`        | `10`        | Window duration in seconds         |
| `DRL_REDIS_HOSTNAME`         | `drl-redis` | Redis hostname (required)          |
| `DRL_REDIS_PORT`             | `6379`      | Redis port (1–65535)               |
| `DRL_REDIS_PASSWORD`         | empty       | Redis password (optional)          |
| `DRL_READ_HEADER_TIMEOUT_MS` | `500`       | HTTP read request header timeout   |
| `DRL_READ_TIMEOUT_MS`        | `500`       | HTTP read request timeout          |
| `DRL_TIMEOUT_MS`             | `1000`      | HTTP response timeout              |

## Running the Application

### Prerequisites

- Docker
- Docker Compose
- Make

### Start

```bash
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
{"success":true}
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

## TODO

- [ ] add GitHub CI pipeline
- [ ] add observability
- [ ] add HTTP rate limit headers
- [ ] additional rate limit algorithms: sliding window and token bucket
- [ ] update Retry-After to return remaining window
- [ ] add health check endpoint
- [ ] add request logging middleware
- [ ] add support for API based client key
- [ ] increase middleware test coverage

## Commit Signature Verification

All commits from 2026-05-25 onward are signed using SSH keys backed by a YubiKey
FIDO2 hardware security key. Each signing operation requires physical presence
on the hardware device. Commits predating this policy are unsigned.

### Verifying commits locally

Configure Git to use the included trust files:

```bash
git config gpg.ssh.allowedSignersFile .allowed_signers
git config gpg.ssh.revocationFile .revoked_signers
```

Verify a specific commit:

```bash
git verify-commit <hash>
```

Verify the full log:

```bash
git log --show-signature
```

### Key rotation and revocation

The `.allowed_signers` file lists all currently trusted public keys. The
`.revoked_signers` file lists keys that must never be trusted, regardless of the
date of the commit they signed. Both files are updated and committed when keys
are added, rotated, or revoked.

In the event of a key compromise, the affected key will be removed from GitHub
and added to `.revoked_signers`. A signed notice commit will be pushed to this
repository identifying the old and new key fingerprints and the date from which
the old key must be considered untrusted.

Public keys for this account are discoverable at:
`https://github.com/snacksforus.keys`
