# Lopor - Commercial-Grade Enterprise AI Workspace

**Lopor** is a high-performance, modular-monolith AI productivity platform designed for high-velocity teams, researchers, and engineers. It combines document editing, RAG knowledge graph retrieval, real-time AI assistant streaming, and autonomous task workflows in a unified, linear-inspired interface.

---

## Architecture At A Glance

```
                         +-----------------------------------+
                         |    Next.js 14/15 App Router UI    |
                         |   (TypeScript, Tailwind, Radix)   |
                         +-----------------+-----------------+
                                           |
                                   HTTP / SSE / WebSocket
                                           |
                                           v
                         +-----------------+-----------------+
                         |    Go Fiber API Modular Monolith  |
                         +-----------------+-----------------+
                                           |
               +---------------------------+---------------------------+
               |                                                       |
               v                                                       v
  +------------+------------+                             +------------+------------+
  | PostgreSQL 16 + pgvector|                             |       Redis 7 Cache |
  | (ACID Relational + Vector)|                             | (Sessions / Rate Limits)|
  +-------------------------+                             +-------------------------+
```

---

## Core Capabilities

- 🤖 **Multi-Provider AI Gateway**: Real-time Server-Sent Events (SSE) streaming for OpenAI, Anthropic, and Ollama compatible endpoints.
- ⚡ **Native pgvector RAG Engine**: High-speed hybrid cosine similarity & lexical search directly in PostgreSQL with citations.
- 📄 **Rich Text Workspace**: Nested documents, folders, tags, version history, and real-time auto-saving.
- 🔐 **Enterprise Security**: JWT rotation, Argon2id password hashing, granular RBAC (Owner/Admin/Member/Viewer), audit logs, and rate limiting.
- 📦 **Modular Monolith in Go**: Built for clean domain isolation and ultra-fast microsecond inter-module execution.

---

## Documentation Blueprint

Full enterprise architectural blueprints can be found in `docs/architecture/`:
- [01-system-overview.md](file:///c:/Users/dhana/My/projects/lopor/docs/architecture/01-system-overview.md): System design, Clean Architecture, and modular monolith structure.
- [02-database-schema.md](file:///c:/Users/dhana/My/projects/lopor/docs/architecture/02-database-schema.md): Complete PostgreSQL & `pgvector` DDL with 18 entities and indexes.
- [03-api-specification.md](file:///c:/Users/dhana/My/projects/lopor/docs/architecture/03-api-specification.md): REST endpoints and SSE streaming protocol.
- [04-ai-rag-pipeline.md](file:///c:/Users/dhana/My/projects/lopor/docs/architecture/04-ai-rag-pipeline.md): Document parsing, chunking, embeddings, and hybrid search.
- [05-security-and-devops.md](file:///c:/Users/dhana/My/projects/lopor/docs/architecture/05-security-and-devops.md): RBAC, JWT rotation, rate limiting, and Docker topology.

---

## Quick Start (Development Setup)

### Prerequisites
- Docker & Docker Compose
- Go 1.22+
- Node.js 20+

```bash
# 1. Start database, cache & monitoring services
docker-compose up -d

# 2. Run backend migrations & server
cd apps/backend
go run cmd/server/main.go

# 3. Start Next.js frontend
cd apps/frontend
npm run dev
```

---

## License
Commercial & Enterprise Reserved.
