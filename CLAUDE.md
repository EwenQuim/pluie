# Pluie

A Go wiki server that serves Obsidian vaults as searchable websites.

## Build & Run

```bash
go build -o pluie .          # Build binary
go run . -path ./vault        # Run locally
go run . -path ./vault -mode static -output ./dist  # Static site generation
go test ./...                 # Run all tests
go test -race ./...           # Run with race detector (important for concurrency)
```

## Project Structure

```
main.go              # Entry point, signal handling, graceful shutdown
server.go            # HTTP server, routes, SSE handlers (Fuego framework)
explorer.go          # Walks vault directory, parses markdown + frontmatter
watcher.go           # fsnotify file watcher for live reload
chat.go              # Chat client initialization (Ollama/Mistral/OpenAI)
weaviate.go          # Weaviate vector store initialization for semantic search
embeddings.go        # Embedding tracking, VectorStore interface, EmbeddingsManager
embedding_progress.go # SSE progress tracking for embedding operations
static.go            # Static site generation

config/              # Configuration loading (env vars, CLI flags, defaults)
engine/              # Core logic: search, tags, tree, backreferences, slugs
model/               # Note data model
template/            # Gomponents HTML templates (SSR, no client templates)
static/              # Embedded static assets (CSS, JS, images)
src/                 # Tailwind CSS source
```

## Architecture

- **HTTP framework**: [Fuego](https://github.com/go-fuego/fuego) (wraps net/http with OpenAPI)
- **Templates**: [Gomponents](https://github.com/maragudk/gomponents) (type-safe Go HTML)
- **AI/Search**: [LangChain Go](https://github.com/tmc/langchaingo) + Weaviate vector store
- **Interactivity**: HTMX + SSE for streaming search results and embedding progress

## Conventions

- Prefer `fuego.Get` (typed handlers) over `fuego.GetStd` (raw `http.HandlerFunc`). Only use `GetStd` when you need direct `http.ResponseWriter` access (SSE streaming, etc.)
- Use named struct types for API responses (not `map[string]string`) so OpenAPI gets proper type declarations. Tag routes with `option.Tags("...")` for OpenAPI grouping.

## Key Invariants

- `path.Join` is for URL slugs/paths; `filepath.Join` is for filesystem operations
- Embedding model must not change without clearing the tracking file (validated on load)
- Notes are private by default; `public: true` frontmatter or `PUBLIC_BY_DEFAULT=true` required
- Config priority: CLI flags > environment variables > defaults
- Embeddings are lazy-loaded on first search access, not on startup
