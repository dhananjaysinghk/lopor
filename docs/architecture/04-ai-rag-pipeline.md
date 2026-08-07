# Production RAG Engine & Vector Pipeline Architecture

## 1. End-to-End RAG Pipeline Architecture

```
[Raw Document] -> [Format Parser] -> [Text Clean & Normalize] -> [Recursive Chunker]
                                                                        |
                                                                        v
[pgvector DB] <--- [Metadata Enriched Vector] <--- [Embedding Engine (text-embedding-3-small)]
      |
      +---> [User Search Query] -> [Query Embedding] -> [HNSW Vector Similarity Search]
                                                                |
                                                                v
[LLM Streaming Output] <--- [Augmented Context Prompt] <--- [Reranker & Citation Builder]
```

---

## 2. Text Ingestion & Chunking Specification

### Ingestion Strategy
- **Supported Formats**: `.pdf`, `.docx`, `.md`, `.txt`, `.csv`, `.json`, `.py`, `.js`, `.ts`, `.go`.
- **Text Cleaners**: Remove null bytes, normalize Unicode canonical decompositions, preserve markdown structures.

### Recursive Character Chunking Parameters
```go
type ChunkConfig struct {
    ChunkSize    int      // 512 tokens (~2000 characters)
    Overlap      int      // 64 tokens (~250 characters)
    Separators   []string // ["\n\n", "\n", ". ", " ", ""]
}
```
- **Chunk Metadata Attributes**:
  - `document_id`: Source document UUID
  - `workspace_id`: Multi-tenant isolation scope
  - `chunk_index`: Sequence position
  - `page_number`: Original page (if PDF)
  - `token_count`: Pre-calculated token length using `tiktoken-go`

---

## 3. Embedding & Vector Search (pgvector)

### Embedding Engine
- Default Model: `text-embedding-3-small` (1536 dimensions) or local `bge-small-en-v1.5` fallback.
- Vector Batching: Embeddings generated in parallel batches of 50 chunks via Goroutines.

### Cosine Similarity Query (SQL)
```sql
SELECT 
    id, 
    chunk_text, 
    metadata, 
    1 - (embedding <=> $1) AS similarity_score
FROM embeddings
WHERE workspace_id = $2
  AND (metadata->>'is_archived')::boolean = false
ORDER BY embedding <=> $1
LIMIT $3;
```

---

## 4. Hybrid Search & Reranking

To maximize precision, Lopor combines **dense vector similarity** with **sparse lexical (BM25/pg_trgm)** keyword search:

1. **Vector Search Score ($S_v$)**: Cosine similarity value $[0.0, 1.0]$.
2. **Text Search Score ($S_t$)**: PostgreSQL `ts_rank_cd` normalized rank.
3. **Combined Hybrid Score**: 
   $$S_{final} = \alpha \cdot S_v + (1 - \alpha) \cdot S_t \quad (\text{Default } \alpha = 0.7)$$
4. **Citation Generation**: Extracted context chunks inject precise references `[Source: File.pdf, Page 3]` into system prompts.
