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

---

## 3. Advanced Enterprise Intelligence Modules (Phases 51–70)

### Real-Time Streaming Token & Cost Calculator (`/api/v1/cost`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/estimate` | Estimate input/output tokens and cost across LLM models | Yes |

### Multi-Agent Consensus & Debate Engine (`/api/v1/workspaces/:wsId/consensus`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/debate` | Multi-persona adversarial consensus debate and synthesis | Yes |
| `POST` | `/quick-vote` | Rapid multi-model voting agreement protocol | Yes |

### Neural Text-to-Speech (TTS) & Audio Narration (`/api/v1`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/voice/tts/voices` | List available neural TTS voices and accents | No |
| `POST` | `/workspaces/:wsId/voice/tts/synthesize` | Convert text to speech audio waveforms | Yes |
| `POST` | `/workspaces/:wsId/voice/tts/document-brief` | Generate spoken executive document briefing | Yes |

### Semantic Document Deduplication & Clustering (`/api/v1/workspaces/:wsId/dedup`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/analyze` | Identify near-duplicate documents and semantic clusters | Yes |
| `POST` | `/merge` | Execute canonical consolidation and redirect links | Yes |

### Event-Driven Workflow Automation & Orchestrator (`/api/v1/workspaces/:wsId/orchestrator`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/workflows` | Register event-driven workflow definition | Yes |
| `GET` | `/workflows` | List workspace workflow pipelines | Yes |
| `POST` | `/dispatch` | Trigger asynchronous workflow execution | Yes |
| `GET` | `/runs/:runId` | Poll workflow execution state and node logs | Yes |

### Multi-Lingual Document Translation & Localization (`/api/v1/workspaces/:wsId/translation`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/languages` | Supported ISO-639 locale codes and capabilities | Yes |
| `POST` | `/translate` | Format-preserving multi-lingual text/markdown translation | Yes |
| `POST` | `/localize-document` | Complete document batch localization pipeline | Yes |

### Red-Team Security Fuzzer & Prompt Resilience (`/api/v1/workspaces/:wsId/redteam`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/fuzz-prompt` | Execute automated prompt adversarial jailbreak fuzzing | Yes |
| `POST` | `/assess-resilience` | Comprehensive prompt safety resilience grading report | Yes |

### Vector Index Optimization & HNSW Calibration (`/api/v1/workspaces/:wsId/rag/index`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/calibrate` | Dynamic tuning of HNSW `m`, `ef_construction`, `ef_search` | Yes |
| `POST` | `/optimize` | Zero-downtime compaction and VACUUM/REINDEX plan generation | Yes |

### SRE Incident Postmortem & Root Cause Analysis (`/api/v1/workspaces/:wsId/incidents`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/templates` | List blameless postmortem section templates | Yes |
| `POST` | `/analyze-rca` | Five-Whys log-driven root cause failure analysis | Yes |
| `POST` | `/generate-postmortem` | Synthesize complete blameless postmortem report | Yes |

### Knowledge Federation & Semantic Linker (`/api/v1/workspaces/:wsId/federation`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/links` | List cross-workspace semantic federation links | Yes |
| `POST` | `/links` | Establish federated entity link between workspaces | Yes |
| `POST` | `/query` | Execute cross-workspace federated RAG query | Yes |

### AI Synthetic Dataset & Benchmark Generator (`/api/v1/workspaces/:wsId/eval/datasets`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/templates` | List benchmark schema templates (MMLU, QA, CodeGen) | Yes |
| `POST` | `/generate` | Synthesize domain-specific evaluation benchmarks | Yes |

### Data Lineage & RAG Provenance Tracking (`/api/v1/workspaces/:wsId/lineage`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/record` | Capture citation hash and chunk transformation event | Yes |
| `GET` | `/trace` | Trace AI completion provenance back to raw document origins | Yes |
| `GET` | `/graph` | Get DAG graph topology of knowledge provenance | Yes |

### Semantic Intent Router & Zero-Shot Classifier (`/api/v1/workspaces/:wsId/router`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/intents` | List platform capability routes and exemplars | Yes |
| `POST` | `/intents` | Register custom workspace intent and target route | Yes |
| `POST` | `/classify` | Sub-5ms zero-shot query classification and endpoint routing | Yes |

### Autonomous Reflection & Self-Correction (`/api/v1/workspaces/:wsId/ai`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/reflect` | Evaluate 4-D critique (Faithfulness, Code, Constraints, Completeness) | Yes |
| `POST` | `/self-correct` | Autonomous iterative reflection loop until threshold satisfied | Yes |

### Long-Term Episodic Memory & Profile Consolidation (`/api/v1/workspaces/:wsId/memory`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/extract` | Extract episodic memories and preferences from conversation | Yes |
| `POST` | `/recall` | Contextual recall with recency decay and prompt injection block | Yes |
| `GET` | `/items` | List user memories filtered by category | Yes |
| `DELETE` | `/items/:memoryId` | Delete specific memory item | Yes |
| `POST` | `/consolidate` | Deduplicate and prune decayed knowledge | Yes |
| `DELETE` | `/purge` | Permanent GDPR Right-to-be-Forgotten memory wipe | Yes |

### Enterprise AI Guardrails & Security Firewall (`/api/v1/workspaces/:wsId/guardrails`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/inspect` | Bi-directional prompt/output firewall inspection | Yes |
| `GET` | `/profile` | Retrieve workspace guardrail thresholds and custom rules | Yes |
| `PUT` | `/profile` | Update workspace severity thresholds and secret masking | Yes |

### Hierarchical GraphRAG & Knowledge Triplet Engine (`/api/v1/workspaces/:wsId/graphrag`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/extract` | Extract entity triplets and detect community clusters | Yes |
| `POST` | `/search/local` | 1-hop and 2-hop entity neighborhood subgraph traversal | Yes |
| `POST` | `/search/global` | Cross-community thematic architectural synthesis | Yes |
| `GET` | `/visual` | Force-directed graph topology schema (D3/Cytoscape) | Yes |

### Multi-Provider Resilience & Circuit Breaker (`/api/v1/workspaces/:wsId/resilience`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/pools/:poolId` | Provider health, circuit states (CLOSED/OPEN/HALF_OPEN), latencies | Yes |
| `POST` | `/pools` | Configure multi-provider ensemble and balance strategies | Yes |
| `POST` | `/execute` | Resilient completion with automatic cascading failover | Yes |

### Human-in-the-Loop Policy & Intervention Gate (`/api/v1/workspaces/:wsId/hitl`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/policies` | Register/update HITL action boundaries and approval timeout policies | Yes |
| `GET` | `/policies` | List active workspace intervention policies | Yes |
| `POST` | `/evaluate` | Evaluate agent action against policy for automated/manual approval | Yes |
| `POST` | `/requests/:requestId/decide` | Submit human decision (APPROVED, REJECTED, AMENDED) | Yes |
| `GET` | `/requests/pending` | List pending actions requiring human approval with risk ratings | Yes |

### Conversational Disambiguation & Clarification (`/api/v1/workspaces/:wsId/clarification`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/analyze` | Evaluate prompt ambiguity and synthesize clarifying options | Yes |
| `POST` | `/resolve` | Resolve ambiguity and synthesize augmented disambiguated prompt | Yes |

### AI Evaluation Harness & LLM-as-a-Judge (`/api/v1/workspaces/:wsId/judge`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/evaluate` | Score model output against rubric dimensions with normalized 0-100 score | Yes |
| `POST` | `/pairwise` | Blind pairwise comparison (Model A vs Model B) with rationale | Yes |
| `GET` | `/rubrics` | List standard workspace evaluation rubrics | Yes |

### Content-Aware Document Chunking & Parser (`/api/v1/workspaces/:wsId/chunking`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/process` | Content-aware recursive chunking with breadcrumbs and key phrases | Yes |

### Adaptive Semantic Vector Query Cache (`/api/v1/workspaces/:wsId/cache`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/lookup` | Approximate cosine vector cache lookup with token savings | Yes |
| `POST` | `/store` | Store query and completion with 64-dim unit embedding | Yes |
| `GET` | `/telemetry` | Retrieve hit rate, tokens saved, and financial USD savings | Yes |
| `DELETE` | `/invalidate` | Purge cached entries for the workspace | Yes |

### Multi-Agent Simulation Sandbox & Roleplay Arena (`/api/v1/workspaces/:wsId/simulation`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/create` | Initialize multi-agent scenario with preset or custom personas | Yes |
| `GET` | `/sessions` | List active and past simulation sessions | Yes |
| `GET` | `/:sessionId` | Retrieve session state, personas, turns, and convergence score | Yes |
| `POST` | `/:sessionId/step` | Advance simulation turn with optional human-in-the-loop input | Yes |
| `POST` | `/:sessionId/run` | Execute simulation autonomously to terminal state | Yes |
| `GET` | `/:sessionId/report` | Compile executive report with action items and sentiment curve | Yes |

### Autonomous Workflow Self-Healing & Fallback (`/api/v1/workspaces/:wsId/selfheal`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/diagnose` | Diagnose execution fault and recommend optimal repair strategy | Yes |
| `POST` | `/remediate` | Execute multi-phase autonomous recovery pipeline | Yes |
| `GET` | `/telemetry` | Retrieve workspace resilience telemetry and success rate | Yes |

### Continuous Prompt Auto-Optimizer & Few-Shot Synthesizer (`/api/v1/workspaces/:wsId/promptopt`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/start` | Start multi-strategy prompt optimization tournament | Yes |
| `GET` | `/jobs` | List prompt optimization jobs | Yes |
| `GET` | `/:jobId` | Retrieve optimization tournament state and candidate variants | Yes |
| `GET` | `/:jobId/report` | Compile executive debrief with visual diff and rollout guidance | Yes |

### Temporal Knowledge Graph & Entity State Evolution (`/api/v1/workspaces/:wsId/temporalkg`)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/facts` | Ingest temporal relation fact with automated conflict superseding | Yes |
| `POST` | `/query` | Execute point-in-time time-travel or temporal interval query | Yes |
| `GET` | `/entities` | List registered temporal entities across categories | Yes |
| `GET` | `/entities/:entityId/evolution` | Retrieve full chronological mutation trajectory for an entity | Yes |


