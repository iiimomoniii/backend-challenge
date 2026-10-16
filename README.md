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

## Authentication selection

Choose the authentication method with `AUTH_METHOD` in your environment or `.env` file. The application defaults to JWT when `AUTH_METHOD` is not set:

```env
# Use the built-in JWT login flow (default)
AUTH_METHOD=jwt

# Or validate bearer tokens through OAuth 2.0 introspection
# AUTH_METHOD=oauth2
```

After changing `.env`, load it again before starting the API:

```sh
source .env
go run ./cmd/api
```

The startup log prints the selected method, for example `authentication method: jwt` or `authentication method: oauth2 token introspection`. With JWT, use `POST /login` to get a token. With OAuth2, the API does not expose `POST /login` or issue tokens; obtain a bearer token from the authorization server and send it to the API.

Configure OAuth2 mode with `OAUTH2_INTROSPECTION_URL`, `OAUTH2_CLIENT_ID`, and `OAUTH2_CLIENT_SECRET`; `OAUTH2_TIMEOUT` defaults to `5s`. The introspection endpoint must use HTTPS in normal configurations and return an active token with a `sub` or `username` value. For local development only, `OAUTH2_ALLOW_INSECURE_HTTP=true` permits HTTP URLs on loopback addresses (`localhost`, `127.0.0.1`, or `::1`). The optional `email` and `exp` values are copied into the request auth context. The verifier sends the token using the RFC 7662 form-encoded introspection request and authenticates as an OAuth client.


### Local OAuth2 mock

A lightweight mock OAuth2 server is available for trying the OAuth2 flow without external credentials. Enable it with `OAUTH2_LOCAL_MOCK=true`; it runs in the same `go run ./cmd/api` process on `127.0.0.1:8090`, supports the `client_credentials` grant at `/token`, and checks access tokens at `/introspect`. It issues the configured test token (default `local-test-token`) for the configured OAuth2 client and maps it to a demo user. This is a development simulation, not a production authorization server or an interactive user login flow.

In Postman, select **OAuth 2.0** and configure **Client Credentials** with Access Token URL `http://127.0.0.1:8090/token`, Client ID `local-mock-client`, Client Secret `local-mock-secret`, and client authentication as **Send as Basic Auth header**. Select **Get New Access Token**, then **Use Token**.

For local mock testing, set these values in `.env` (use the same client credentials for the mock and introspection):

```env
AUTH_METHOD=oauth2
OAUTH2_LOCAL_MOCK=true
OAUTH2_INTROSPECTION_URL=http://127.0.0.1:8090/introspect
OAUTH2_ALLOW_INSECURE_HTTP=true
OAUTH2_CLIENT_ID=local-mock-client
OAUTH2_CLIENT_SECRET=local-mock-secret
```

Then start both the API and mock with one command:

```sh
source .env
go run ./cmd/api
```

Or call an authenticated endpoint with `Authorization: Bearer local-test-token`, for example:

```sh
curl -H "Authorization: Bearer local-test-token" http://localhost:8085/users
```

Any other token is reported inactive and receives `401`. The mock credentials and loopback HTTP exception are for local development only.
