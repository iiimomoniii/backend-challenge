# Backend Challenge

Backend implementation for the 7Solutions Backend Challenge.

The project implements a User Management API with JWT authentication,
MongoDB persistence, middleware, background worker, graceful shutdown,
and a separate design proposal for the Lottery Search System.

## Tech Stack

- Go
- Standard `net/http`
- MongoDB
- MongoDB Go Driver
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
│       └── mongo/              # MongoDB repository
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

docker-compose.yml              # Local MongoDB

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
MongoDB / JWT / bcrypt