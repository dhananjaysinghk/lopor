# System Architecture & Modular Monolith Overview

## 1. Executive Summary

**Lopor** is a commercial-grade, enterprise AI Workspace built to empower knowledge workers, engineers, researchers, and teams to think, write, collaborate, and automate workflows using AI. 

The architecture is designed as a **Modular Monolith in Go**, paired with a **Next.js (App Router)** frontend and a high-performance **PostgreSQL + pgvector** storage engine. This architecture guarantees single-binary deployment simplicity and low operational overhead while maintaining strict domain separation so individual modules can be extracted into microservices when scaling demands require it.

---

## 2. High-Level System Architecture

```
                                +----------------------------------+
                                |  Next.js 14/15 Frontend Client   |
                                |   (React, TypeScript, Tailwind)  |
                                +-----------------+----------------+
                                                  |
                                            HTTPS / WSS / SSE
                                                  |
                                                  v
                                +-----------------+----------------+
                                |        Go API Gateway / Router   |
                                |     (Fiber Framework / Middleware)|
                                +-----------------+----------------+
                                                  |
      +----------------------+--------------------+----------------------+----------------------+
      |                      |                    |                      |                      |
      v                      v                    v                      v                      v
+-----+------+        +------+-----+        +-----+------+        +------+-----+        +-------+----+
| Auth & RBAC|        | Workspace  |        | AI Gateway |        | RAG Engine |        | Storage &  |
|   Module   |        |  & Docs    |        | & Streaming|        | & Vector   |        | Async Jobs |
+-----+------+        +------+-----+        +-----+------+        +------+-----+        +-------+----+
      |                      |                    |                      |                      |
      +----------------------+--------------------+----------------------+----------------------+
                                                  |
                                 +----------------+----------------+
                                 |                                 |
                                 v                                 v
                     +-----------+-----------+         +-----------+-----------+
                     | PostgreSQL 16 + pgx   |         |      Redis 7 Cache    |
                     |  (Relational + Vector)|         | (Sessions/Rate Limits)|
                     +-----------------------+         +-----------------------+
```

---

## 3. Modular Monolith Architecture Rationale

### Why Modular Monolith?
1. **Low Operational Overhead**: No distributed tracing setup overhead, network latency, or split-brain deployments during early production phases.
2. **Strict Domain Boundaries**: Each domain module (`auth`, `workspace`, `document`, `ai`, `rag`, `agent`, `notification`) owns its database tables, models, services, and HTTP handlers.
3. **In-Memory Inter-Module Communication**: Modules invoke each other via Go interfaces or an internal Go event bus (`pkg/eventbus`), enabling microsecond latency for cross-domain calls.
4. **Microservice Extractability**: Each module communicates strictly through defined Go contracts, allowing seamless extraction into standalone gRPC/REST microservices when needed.

---

## 4. Layered Clean Architecture within Modules

Each module follows clean architecture boundaries:

```
apps/backend/internal/domain/<module>/
├── handler.go      # HTTP / SSE Delivery Layer (Request parsing, status codes, response rendering)
├── service.go      # Core Business Logic Layer (Use cases, validation, transaction boundaries)
├── repository.go   # Data Access Layer (PostgreSQL pgx queries, Redis cache lookups)
├── model.go        # Domain Entities & DTO Structs
└── handler_test.go # Unit & Integration Tests
```

### Flow of Execution:
1. **HTTP Request** received by Fiber Router -> Passed through Auth / Rate-Limit Middleware.
2. **Handler**: Validates request parameters & JSON body schema.
3. **Service**: Enforces domain rules, checks RBAC permissions, coordinates repositories & AI providers.
4. **Repository**: Executes optimized SQL query (pgx pool) or Redis cache lookup.
5. **Response**: Formatted standard JSON or SSE event stream returned to frontend.

---

## 5. Technology Stack Rationale

| Layer | Technology | Rationale |
| :--- | :--- | :--- |
| **Frontend Framework** | Next.js 14/15 App Router | Server Components for instant page rendering, client components for interactive rich-text editor & real-time chat. |
| **State & Fetching** | TanStack Query v5 + Zustand | Optimistic UI updates, caching, background polling, and predictable local component state. |
| **Styling & UI** | Tailwind CSS + Framer Motion + Radix UI | Modern dark/light theme tokens, fluid animations, accessible modal & dialog primitives. |
| **Backend Language** | Go (1.22+) | Ultra-fast execution, low memory footprint, native concurrency (goroutines) for streaming & async RAG processing. |
| **HTTP Framework** | Go Fiber v2 | Express-like developer ergonomics with zero-allocation HTTP engine built on `fasthttp`. |
| **Database** | PostgreSQL 16 + `pgvector` | Native relational ACID guarantees + high-performance HNSW vector indexing without separate vector DB infrastructure. |
| **Caching & PubSub** | Redis 7 | High-speed session store, distributed rate limiting, and SSE message broadcast channel. |
| **File Storage** | AWS S3 / Cloudflare R2 | Zero-egress cost object storage for raw documents, previews, and user assets. |
