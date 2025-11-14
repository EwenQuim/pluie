# Pluie 🌧️

![](./static/pluie.webp)

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

**Environment variables:**

```bash
SITE_TITLE="My Knowledge Base"
PUBLIC_BY_DEFAULT=false    # Notes are private by default
HOME_NOTE_SLUG=Index       # Landing page
PORT=9999
```

**Privacy control:**

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
