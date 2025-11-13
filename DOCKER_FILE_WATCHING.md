# File Watching in Docker

Pluie includes built-in file system watching that automatically detects changes to your markdown files and reloads the content without restarting the server.

## Configuration

### Default Behavior

File watching is **enabled by default**. No configuration needed - just mount your vault and start the container:

```bash
docker compose up
```

Your changes will be automatically detected and reflected in ~1 second.

### How It Works

1. **Volume Mount**: Your local vault directory is mounted to `/vault` in the container
2. **fsnotify**: The Go fsnotify library monitors the mounted directory for changes
3. **Auto-reload**: When files change, the server automatically reloads content within ~1 second (including 500ms debounce)

## Docker Compose Setup

Your `docker-compose.yml` should include:

```yaml
services:
  pluie:
    image: ewenquim/pluie:latest
    volumes:
      - ./vault:/vault  # Mount your local vault
    ports:
      - "9999:9999"
```

## Platform-Specific Considerations

### Linux
✅ **Fully supported** - Native inotify events work perfectly with Docker volumes.

### macOS (Docker Desktop)
✅ **Supported with slight delay** 
- Uses osxfs/VirtioFS for volume mounts
- File events propagate correctly but may have 1-2 second delay
- **Recommendation**: Use `:cached` or `:delegated` mount options for better performance:
  ```yaml
  volumes:
    - ./vault:/vault:cached
  ```

### Windows (Docker Desktop)
✅ **Supported with delay**
- Uses Hyper-V or WSL2 for volume mounts
- File events work but may have slight delay
- **WSL2 is recommended** over Hyper-V for better performance

## Troubleshooting

### Changes Not Detected

**Check the logs:**
```bash
docker compose logs -f pluie
```

You should see:
```
INFO Added directories to watcher count=X root=/vault
INFO File watcher started path=/vault
```

When you edit a file, you should see:
```
INFO File change detected file=/vault/your-file.md op=WRITE
INFO Reloading notes due to file changes
INFO Notes reloaded successfully
```

### No watcher logs appearing

1. **Verify volume mount:**
   ```bash
   docker compose exec pluie ls -la /vault
   ```

2. **Check if watch is enabled:**
   ```bash
   docker compose exec pluie ps aux | grep pluie
   ```
   Should show: `./main -path /vault -watch true`

3. **Test with simple file change:**
   ```bash
   echo "test" >> vault/test.md
   docker compose logs --tail=20 pluie
   ```

### Performance Issues

If file watching causes high CPU or many events:

1. **Exclude large directories** - Add a `.pluie` file in directories you want to skip
2. **Reduce debounce time** - Edit watcher.go if needed (currently 500ms)
3. **Check for file editor behavior** - Some editors create many temporary files

## Disabling File Watching

If you need to disable file watching (e.g., for very large vaults or performance reasons):

### In docker-compose.yml

Uncomment or add the `command` line:

```yaml
services:
  pluie:
    image: ewenquim/pluie:latest
    volumes:
      - ./vault:/vault
    command: ./main -path /vault -watch=false  # Disable file watching
```

### In plain Docker

```bash
docker run -v ./vault:/vault ewenquim/pluie ./main -path /vault -watch=false
```

### Custom Dockerfile

```dockerfile
FROM ewenquim/pluie:latest
CMD ["./main", "-path", "/vault", "-watch=false"]
```

## Development Mode

For active development with hot-reload of the application itself (not just markdown files), use:

```bash
docker compose -f docker-compose.dev.yml up --build --watch
```

This watches for:
- **Code changes** → Rebuilds the container
- **Vault changes** → Syncs files to container (file watcher detects them)

## Technical Details

### What Gets Watched

- All directories recursively under `/vault`
- Excludes hidden directories (starting with `.` except the root)
- Excludes `.git`, `node_modules`

### Detected Events

- `CREATE` - New files
- `WRITE` - File modifications  
- `REMOVE` - File deletions
- `RENAME` - File/directory renames
- `CHMOD` - Permission changes (on some systems)

### Reload Process

1. File change detected → Start debounce timer (500ms)
2. Additional changes reset the timer
3. Timer expires → Reload all notes
4. Parse markdown, build tree, update server
5. New requests serve updated content

### Resource Usage

- **Memory**: ~1-5MB for watcher goroutine
- **CPU**: Negligible when idle, brief spike during reload
- **I/O**: Minimal - only reads changed files

## Best Practices

1. **Use specific vault paths** - Don't mount entire home directory
2. **Group related files** - Use folders for better organization
3. **Avoid rapid changes** - Debounce handles this, but try to save files once
4. **Monitor logs initially** - Verify file watching works correctly
5. **Test on your platform** - Verify delay is acceptable on Mac/Windows

## Related Configuration

See also:
- `docker-compose.yml` - Production setup
- `docker-compose.dev.yml` - Development setup  
- `Dockerfile` - Container configuration
- `watcher.go` - File watching implementation
