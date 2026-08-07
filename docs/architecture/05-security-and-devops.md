# Security Architecture, RBAC & DevOps Blueprint

## 1. Security Architecture & Threat Mitigation

### Authentication & Token Lifecycle
1. **Access Token (JWT)**: Short-lived (15 minutes). Contains `user_id`, `email`, `role`, and active `workspace_id`. Signed with RSA-256 / HS256.
2. **Refresh Token**: Long-lived (7 days), stored in HTTP-Only, Secure, `SameSite=Lax` cookies, hashed with SHA-256 in the PostgreSQL database.
3. **Password Security**: Hashed using **Argon2id** (memory=64MB, iterations=3, parallelism=2).

### Role-Based Access Control (RBAC) Hierarchy
```
System Level: SuperAdmin -> Admin -> User
Workspace Level: Workspace Owner -> Workspace Admin -> Member -> Viewer
```
- **Middleware Guard**: Every API route verifies workspace membership and permissions before passing execution to domain handlers.

### Rate Limiting & Abuse Prevention
- **Sliding Window Rate Limiter**: Powered by Redis (`golang-jwt` + `go-redis`).
- Standard API: 100 requests / min / IP.
- Auth endpoints: 5 login attempts / 15 mins / IP.
- AI Chat SSE streams: 20 prompt requests / min / user.

---

## 2. Infrastructure & Containerization (Docker Topology)

### Production Container Layout (`docker-compose.yml`)
1. **Frontend App (`lopor-web`)**: Next.js Node container (Port 3000).
2. **Backend API (`lopor-api`)**: Compiled static Go binary container (Port 8080).
3. **Database (`lopor-db`)**: `pgvector/pgvector:pg16` PostgreSQL vector database (Port 5432).
4. **Cache & Bus (`lopor-redis`)**: Redis 7 Alpine image (Port 6379).
5. **Monitoring (`prometheus` & `grafana`)**: Metrics scraper and dashboard viz (Ports 9090 / 3000).

---

## 3. Security Checklist

- [x] TLS 1.3 enforced on all client connections.
- [x] CORS strict origin headers matched against frontend domain.
- [x] SQL injection protection via `pgx` parameterized prepared queries.
- [x] XSS prevention using DOMPurify on frontend markdown rendering.
- [x] CSRF protection using SameSite cookies and Custom Header verification.
- [x] Environment variable secrets strictly loaded from `.env` or AWS KMS.
- [x] Comprehensive immutable Audit Logging for security-sensitive actions.
