# Pluie 🌧️

<p align="center">
  <img src="./static/pluie.webp" alt="Pluie" width="400">
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/EwenQuim/pluie">
    <img src="https://goreportcard.com/badge/github.com/EwenQuim/pluie" alt="Go Report Card">
  </a>
  <a href="https://github.com/go-fuego/fuego">
    <img src="https://img.shields.io/badge/built%20with-Fuego-red" alt="Built with Fuego">
  </a>
</p>

A lightweight markdown wiki server that turns your Obsidian vault into a searchable website. Self-hosted, privacy-focused, and built in Go.

## What it does

Pluie serves your markdown notes as a wiki with full support for `[[wiki links]]`, automatic backreferences, and real-time search. It respects your folder structure and lets you control what's public or private at the note or folder level.

**Core features:**

- Wiki linking with Obsidian-style `[[Note Name]]` syntax and automatic backreferences
- Real-time search with keyboard shortcuts (Cmd/Ctrl + K)
- Granular privacy controls per note or folder
- Collapsible folder tree that mirrors your vault structure
- File watcher automatically reloads notes when they change (enabled by default)
- Server mode with ready-to-use Docker image...
- ...or Static site generation for deploying to GitHub Pages, Netlify, or any static host

The server loads your notes into memory on startup for fast access. When you edit markdown files in your vault, the watcher detects changes and reloads them automatically. Everything runs in a single binary with no external dependencies—CSS and JavaScript are embedded.

## Why Pluie?

If you want something between a static site generator and Obsidian Publish, Pluie might be for you. It's completely free and self-hosted, so your notes never leave your infrastructure. The learning curve is minimal: point it at your vault and it works.

Compared to static generators like Hugo, you don't need to configure anything or learn a templating language. Just write markdown in Obsidian and Pluie handles the rest. You can also generate a static version of your site for deployment anywhere.

## Quick Start

**Docker (using pre-built image):**

```bash
docker pull ghcr.io/ewenquim/pluie:latest
mkdir vault
echo "# Welcome to Pluie" > vault/Index.md
docker run -v $(pwd)/vault:/vault -p 9999:9999 ghcr.io/ewenquim/pluie:latest -path /vault
```

**Docker (building from source):**

```bash
git clone https://github.com/EwenQuim/pluie.git
cd pluie
mkdir vault
echo "# Welcome to Pluie" > vault/Index.md
docker-compose up --build
```

Open `http://localhost:9999` to see your wiki.

**Local:**

```bash
go run . -path ./vault
```

The file watcher runs by default and reloads your notes automatically when they change.

**Static site generation:**

```bash
./pluie -path ./vault -mode static -output ./public
```

This generates a static HTML site in the `./public` folder that you can deploy to GitHub Pages, Netlify, or any static host.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SITE_TITLE` | `Pluie` | Site title displayed in the header |
| `SITE_ICON` | `/static/pluie.webp` | Path to the site icon |
| `SITE_DESCRIPTION` | _(empty)_ | Site description for meta tags |
| `PUBLIC_BY_DEFAULT` | `false` | If `true`, all notes are public unless explicitly private |
| `HOME_NOTE_SLUG` | `Index` | Slug of the note to use as the landing page |
| `HIDE_YAML_FRONTMATTER` | `false` | If `true`, frontmatter is hidden from rendered notes |
| `PORT` | `9999` | HTTP server port |
| `LOG_JSON` | `false` | Enable JSON logging (default: pretty logging) |

### AI / Chat

Pluie supports AI-powered search responses via Ollama (local), Mistral, or OpenAI.

| Variable | Default | Description |
|----------|---------|-------------|
| `CHAT_PROVIDER` | `ollama` | Chat provider: `ollama`, `mistral`, or `openai` |
| `CHAT_MODEL` | `tinyllama` | Model name for the chat provider |
| `OLLAMA_URL` | `http://ollama-models:11434` | Ollama server URL |
| `MISTRAL_API_KEY` | _(empty)_ | Mistral API key (required when using `mistral` provider) |
| `OPENAI_API_KEY` | _(empty)_ | OpenAI API key (required when using `openai` provider) |

### Embeddings / Weaviate

Semantic search uses vector embeddings stored in Weaviate. Without Weaviate, only title and heading search is available.

| Variable | Default | Description |
|----------|---------|-------------|
| `EMBEDDING_PROVIDER` | `ollama` | Embedding provider: `ollama`, `openai`, or `mistral` |
| `EMBEDDINGS_TRACKING_FILE` | `embeddings_tracking.json` | Path to the file tracking which notes have been embedded |
| `WEAVIATE_HOST` | `weaviate-embeddings:9035` | Weaviate server host |
| `WEAVIATE_SCHEME` | `http` | Weaviate connection scheme (`http` or `https`) |
| `WEAVIATE_INDEX` | `Note` | Weaviate index/class name |

Embeddings are created lazily on first search access. By default they use Ollama with `nomic-embed-text`, but you can switch to OpenAI or Mistral embedding models via `EMBEDDING_PROVIDER`.

### Static Mode

Static site generation (`-mode static`) produces HTML files but does not include search or AI features. These require a running server with Weaviate and a chat provider.

### Privacy Control

Control note visibility with frontmatter:

```yaml
---
public: true
title: "My Public Note"
---
```

Or set defaults for entire folders with a `.pluie` file:

```yaml
---
public: false
---
```

## Contributing

Bug reports, feature requests, and pull requests are welcome. Run tests with `go test ./...` and test your changes with `go run . -path ./testdata/test_notes`.

## License

MIT License - see [LICENSE](LICENSE) for details.

Built with [Go](https://golang.org/), [Fuego](https://github.com/go-fuego/fuego), [Gomponents](https://github.com/maragudk/gomponents), and [HTMX](https://htmx.org/).
