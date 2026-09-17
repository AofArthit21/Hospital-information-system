# Agnos Hospital Middleware

Backend assignment implementation for searching patient data through a hospital-scoped middleware API. The service uses Go, Gin, PostgreSQL, Nginx, and Docker Compose.

## What is included

- Staff registration and login with bcrypt password hashes and signed JWTs
- Hospital isolation: the hospital is taken from the validated JWT, never from patient-search input
- Hospital A HIS integration through `GET {his_base_url}/patient/search/{id}`
- Patient cache/upsert in PostgreSQL and search by all fields in the assignment
- PostgreSQL constraints and indexes for hospital-scoped identifiers
- Nginx reverse proxy, rate limiting, container health checks, and graceful shutdown
- Unit tests for positive and negative scenarios on every API
- [API specification](docs/openapi.yaml), [ER diagram](docs/ER_DIAGRAM.md), and [implementation plan](docs/PLAN.md)

## Run with Docker Compose

Docker Desktop (or another Docker daemon) must be running.

```bash
docker compose up --build
```

The API is exposed through Nginx at `http://localhost:8080`. The migration seeds a hospital with code `hospital-a` and HIS URL `https://hospital-a.api.co.th`.

> PostgreSQL only runs files in `/docker-entrypoint-initdb.d` for a new volume. During development, after changing the migration, recreate the local volume with `docker compose down -v` and start again. This deletes local development data.

## Try the API

Create staff:

```bash
curl -X POST http://localhost:8080/staff/create \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123","hospital":"hospital-a"}'
```

Login:

```bash
curl -X POST http://localhost:8080/staff/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123","hospital":"hospital-a"}'
```

Search by national ID (this refreshes the patient from the hospital HIS first):

```bash
curl "http://localhost:8080/patient/search?national_id=1234567890123" \
  -H "Authorization: Bearer <access_token>"
```

Search cached patients by multiple optional criteria:

```bash
curl "http://localhost:8080/patient/search?first_name=Somchai&date_of_birth=1990-01-02" \
  -H "Authorization: Bearer <access_token>"
```

Calling patient search without query parameters returns at most the 100 most recently updated patients belonging to the authenticated staff member's hospital.

## Local development

Copy `.env.example` values into your environment, start PostgreSQL with the migration, then run:

```bash
go run ./cmd/api
go test ./...
go vet ./...
```

To point Hospital A at a local mock HIS, update its stored base URL:

```sql
UPDATE hospitals SET his_base_url = 'http://host.docker.internal:9090' WHERE code = 'hospital-a';
```

## Project structure

```text
cmd/api/                         application entry point
internal/auth/                   JWT creation and validation
internal/config/                 environment configuration
internal/domain/                 entities, filters, and domain errors
internal/his/                    Hospital A HTTP adapter
internal/httpapi/                Gin routes, middleware, and handlers
internal/repository/postgres/    PostgreSQL repositories
internal/service/                authentication and patient use cases
migrations/                      PostgreSQL schema and seed data
deploy/                          Nginx configuration
docs/                            API spec, ER diagram, and plan
```

## Important behavior

- A username is unique within a hospital, so the same username may exist in two hospitals.
- Passwords are never stored or returned in plaintext.
- JWT claims contain staff and hospital IDs; a caller cannot select another hospital in `/patient/search`.
- An ID search calls the HIS, caches a successful response, and then applies all supplied filters in PostgreSQL.
- A HIS `404` falls back to the hospital's local cache (and is empty when no cache entry exists). HIS transport, invalid-response, and 5xx errors yield `502 HIS_UNAVAILABLE`.
- Name filters are case-insensitive partial matches across Thai and English names. Other filters are exact matches (email is case-insensitive).

## Production notes

Use a secrets manager for `JWT_SECRET` and database credentials, terminate TLS in front of Nginx, restrict staff creation to an administrative workflow, and use a versioned migration runner instead of the development-time PostgreSQL init directory.
