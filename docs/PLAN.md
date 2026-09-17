# Development Plan and Architecture

## Scope

The middleware authenticates hospital staff, retrieves a patient by national/passport ID from that staff member's hospital information system (HIS), caches the normalized record, and lets the staff member search cached patient records using the criteria from the assignment.

## Request flow

1. `POST /staff/create` resolves the supplied hospital code, hashes the password with bcrypt, and inserts a hospital-scoped staff record.
2. `POST /staff/login` verifies hospital, username, and password and issues an HMAC-SHA256 JWT containing immutable staff and hospital IDs.
3. JWT middleware validates signature and expiry for `GET /patient/search` and places the trusted IDs in request context.
4. For a national/passport ID query, the patient service requests `/patient/search/{id}` from the hospital's configured HIS.
5. A valid HIS record is normalized and upserted with `(hospital_id, patient_hn)` as the stable cache key.
6. The PostgreSQL repository always starts its query with `WHERE hospital_id = $1`, using the ID from the token, and adds only parameterized filters.

## Design decisions

- **Layered, dependency-inverted packages:** HTTP, use cases, HIS transport, and persistence can be tested or replaced separately.
- **Hospital-scoped JWT claim:** prevents clients from changing hospital scope using request parameters.
- **Database constraints:** staff usernames and patient identifiers are unique per hospital, while foreign keys preserve ownership.
- **HIS adapter:** HTTP status and payload errors are translated to stable domain errors; transport has a configurable timeout.
- **Cache plus local search:** Hospital A only documents lookup by one ID. The cache provides the remaining multi-field search behavior without inventing unsupported upstream endpoints.
- **Bounded result set:** unfiltered and filtered searches return at most 100 records.
- **No leaked internals:** API errors have stable codes and do not return database, password, token-validation, or upstream details.
- **PII-safe access logs:** both Gin and Nginx omit query strings, which may contain patient identifiers and contact details.

## Test strategy

- Handler tests exercise success and failure paths for staff creation, login, and protected patient search.
- Token-backed handler tests prove that hospital scope reaches the service from the JWT.
- HIS adapter tests cover valid normalization, not-found behavior, and invalid upstream payloads.
- `test/e2e/mock-his-nginx.conf` provides a deterministic HIS response for local Docker acceptance tests without sending patient identifiers to an external host.
- `go test ./...` runs without external services; repository correctness is additionally protected by parameterized SQL, schema constraints, and a Compose smoke test when Docker is available.

## Deployment topology

```text
client -> nginx:80 -> Go/Gin API:8080 -> PostgreSQL:5432
                              |
                              +----------> Hospital HIS HTTPS API
```

Only Nginx publishes a host port. PostgreSQL and the Go service stay on the private Compose network.

## Future enhancements

- Administrative authorization and invitations for `/staff/create`
- Refresh-token rotation and server-side token revocation
- Per-hospital HIS credentials stored in a secrets manager
- Audit log for patient searches and personally identifiable information access
- Cursor pagination and PostgreSQL trigram indexes for larger patient volumes
- Resilience policy with safe retries/circuit breaking for HIS calls
