# Pluie - Markdown Wiki Server

Pluie is a Go-based web application that serves Obsidian-style markdown notes with wiki-style linking and live search functionality.

## Features

- **Wiki-style links**: Create links between notes using `[[Note Name]]` syntax
- **Live search**: Real-time search through your notes
- **Markdown rendering**: Full markdown support with syntax highlighting
- **Docker deployment**: Easy containerized deployment
- **Volume persistence**: Your notes are stored persistently in the `/vault` directory

## Docker Setup

### Prerequisites

- Docker and Docker Compose installed on your system

### Quick Start

1. **Clone the repository** (if not already done):

   ```bash
   git clone <repository-url>
   cd pluie
   ```

2. **Add your markdown files** to the `vault/` directory:

   ```bash
   # Your markdown files go here
   vault/
   ├── HOME.md          # Default home page
   ├── notes/
   ├── projects/
   └── ...
   ```

3. **Start the application**:

   ```bash
   docker-compose up --build
   ```

4. **Access the application**:
   Open your browser and navigate to `http://localhost:9999`

### Configuration

The docker-compose.yml file includes the following configuration:

- **Port mapping**: `9999:9999` (host:container)
- **Volume mount**: `./vault:/vault` (your local vault directory is mounted to `/vault` in the container)
- **Environment variables**:
  - `HOME_NOTE_SLUG=HOME` - Sets the default home page to `HOME.md`

### Directory Structure

```
pluie/
├── Dockerfile              # Multi-stage Docker build
├── docker-compose.yml      # Docker Compose configuration
├── vault/                  # Your markdown notes directory
│   ├── HOME.md            # Default home page
│   └── ...                # Your other markdown files
├── main.go                # Application entry point
├── server.go              # Web server implementation
└── ...                    # Other Go source files
```

### Usage

1. **Home Page**: The application will serve the note specified by `HOME_NOTE_SLUG` environment variable (defaults to `HOME.md`)

2. **Navigation**:

   - Browse notes using the sidebar
   - Use the search functionality to find specific notes
   - Click on wiki-style links `[[Note Name]]` to navigate between notes

3. **Adding Notes**:
   - Add new `.md` files to the `vault/` directory
   - The application will automatically discover them on restart
   - Use wiki-style linking to connect your notes

### Docker Commands

#### Production Commands

```bash
# Start the application
docker-compose up

# Start with rebuild
docker-compose up --build

# Run in background
docker-compose up -d

# Stop the application
docker-compose down

# View logs
docker-compose logs -f
```

#### Development Commands with Docker Watch

```bash
# Start development environment with hot reloading
docker compose -f docker-compose.dev.yml up --build --watch

# Start development environment in background with watch
docker compose -f docker-compose.dev.yml up -d --watch

# Stop development environment
docker compose -f docker-compose.dev.yml down

# View development logs
docker compose -f docker-compose.dev.yml logs -f
```

**Docker Watch Features:**

- **Hot Reloading**: Automatically rebuilds and restarts the application when Go source files change
- **File Sync**: Syncs vault directory changes without container restart
- **Air Integration**: Uses Air for fast Go application reloading
- **Development Optimized**: Faster rebuilds with development-focused Dockerfile

### Environment Variables

- `HOME_NOTE_SLUG`: Specifies which note to use as the home page (default: looks for HOME.md, then first alphabetical note)
- `PATH_FLAG`: Internal variable set to `/vault` in the container

### Troubleshooting

1. **Port already in use**: If port 9999 is already in use, modify the port mapping in `docker-compose.yml`:

   ```yaml
   ports:
     - "8080:9999" # Use port 8080 instead
   ```

2. **Notes not appearing**: Ensure your markdown files are in the `vault/` directory and have `.md` extension

3. **Permission issues**: Make sure the `vault/` directory is readable by the Docker container

## Development

To run the application locally without Docker:

```bash
# Install dependencies
go mod download

# Run the application
go run . -path ./vault

# Or build and run
go build -o pluie
./pluie -path ./vault
```

The application will be available at `http://localhost:9999`.
