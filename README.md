# Backend Challenge

Backend implementation for the 7Solutions Backend Challenge.

The project implements a User Management API with JWT authentication,
MongoDB or PostgreSQL persistence, middleware, background worker, graceful shutdown,
and a separate design proposal for the Lottery Search System.

## Tech Stack

- Go
- Standard `net/http`
- MongoDB / PostgreSQL
- MongoDB Go Driver / lib/pq
- JWT
- bcrypt
- Docker / Docker Compose

## Architecture

The project follows a Hexagonal Architecture / Ports and Adapters approach.

```text
cmd/
└── api/
    └── main.go                 # Application entry point / composition root

internal/
├── application/
│   ├── auth/                   # Authentication use cases
│   ├── user/                   # User use cases
│   └── worker/                 # Background workers
│
├── domain/
│   ├── auth/                   # Authentication domain contracts
│   └── user/                   # User entity, repository and domain errors
│
├── infrastructure/
│   ├── auth/                   # JWT implementation
│   ├── hasher/                 # bcrypt implementation
│   └── persistence/
│       ├── mongodb/user/       # MongoDB repository
│       └── postgresql/user/    # PostgreSQL repository
│
└── interface/
    └── http/
        ├── handler/            # HTTP handlers
        ├── middleware/         # HTTP middleware
        ├── request/            # HTTP request models
        ├── response/           # HTTP response models
        └── router.go           # HTTP route configuration

pkg/
└── code/                       # Error codes and localization

docker-compose.yml              # Local MongoDB and PostgreSQL

######################

Dependency Direction

HTTP Handler
     │
     ▼
Application Use Case
     │
     ▼
Domain Interface
     │
     ▼
Infrastructure Adapter
     │
     ▼
MongoDB or PostgreSQL / JWT / bcrypt

## Database selection

MongoDB remains the default. Set `DATABASE_DRIVER=postgresql` to use PostgreSQL.
The local PostgreSQL connection defaults to `postgres://backend_challenge:backend_challenge@localhost:5432/backend_challenge?sslmode=disable`; override it with `POSTGRES_URI`.
The PostgreSQL adapter creates the `users` table on startup with the existing user fields (`id`, `name`, `email`, `password`, `created_at`).

Start the local PostgreSQL service with `docker compose up -d postgresql`.
