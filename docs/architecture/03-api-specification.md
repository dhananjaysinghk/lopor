# REST & SSE Streaming API Specification

## 1. Global API Conventions

- **Base URL**: `/api/v1`
- **Content Type**: `application/json` (unless uploading files or streaming responses)
- **Authentication Header**: `Authorization: Bearer <JWT_ACCESS_TOKEN>`
- **Error Response Format**:
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired authentication token",
    "details": null
  }
}
```

---

## 2. API Endpoints Matrix

### Authentication Module (`/api/v1/auth`)
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `POST` | `/register` | Register new user account | No |
| `POST` | `/login` | User login (returns JWT + set HTTP-Only refresh cookie) | No |
| `POST` | `/logout` | Invalidate current session refresh token | Yes |
| `POST` | `/refresh` | Obtain new access token using refresh token | No |
| `GET` | `/me` | Get current authenticated user details | Yes |
| `POST` | `/forgot-password` | Request password reset email | No |
| `POST` | `/reset-password` | Execute password reset using token | No |

---

### Workspaces & Organizations (`/api/v1/workspaces`)
| Method | Endpoint | Description | Role |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | List all user accessible workspaces | Member |
| `POST` | `/` | Create a new workspace | User |
| `GET` | `/:id` | Get workspace details | Member |
| `PATCH` | `/:id` | Update workspace metadata | Admin/Owner |
| `DELETE` | `/:id` | Delete workspace | Owner |
| `GET` | `/:id/members` | List members in workspace | Member |
| `POST` | `/:id/members/invite` | Send workspace invitation email | Admin/Owner |
| `DELETE` | `/:id/members/:userId` | Remove member from workspace | Admin/Owner |

---

### AI Chat & Streaming Gateway (`/api/v1/workspaces/:wsId/chats`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | List workspace chat sessions | Yes |
| `POST` | `/` | Initialize new chat session | Yes |
| `GET` | `/:chatId` | Fetch chat details & message history | Yes |
| `PATCH` | `/:chatId` | Update chat title / toggle pinned status | Yes |
| `DELETE` | `/:chatId` | Delete chat session | Yes |
| `POST` | `/:chatId/stream` | **Server-Sent Events (SSE)** endpoint for streaming response | Yes |

#### SSE Event Payload Example (`/chats/:chatId/stream`):
```http
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: message_start
data: {"message_id": "msg_987", "role": "assistant"}

event: content_block_delta
data: {"delta": "Here is the "}

event: content_block_delta
data: {"delta": "analyzed document summary..."}

event: citations
data: [{"doc_id": "doc_123", "title": "Quarterly Report", "chunk_id": "chk_45"}]

event: message_end
data: {"usage": {"prompt_tokens": 120, "completion_tokens": 45}}
```

---

### RAG & Vector Search (`/api/v1/workspaces/:wsId/search`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/semantic` | Perform pgvector cosine similarity search | Yes |
| `POST` | `/hybrid` | Combined BM25 text search + Vector embedding search | Yes |

#### Search Body Schema:
```json
{
  "query": "What are our project deadlines for Q3?",
  "top_k": 5,
  "filter_tags": ["project-x"],
  "min_score": 0.75
}
```

---

### Documents & File Ingestion (`/api/v1/workspaces/:wsId/documents`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | List workspace documents | Yes |
| `POST` | `/` | Create new document | Yes |
| `GET` | `/:id` | Get document content and metadata | Yes |
| `PATCH` | `/:id` | Save/Autosave document content | Yes |
| `DELETE` | `/:id` | Soft delete / archive document | Yes |
| `POST` | `/upload-file` | Upload file (PDF, DOCX, Code) & queue background RAG embedding | Yes |
