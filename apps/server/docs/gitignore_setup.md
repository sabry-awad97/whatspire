# Git Ignore Configuration

## Overview

Proper .gitignore setup to exclude binary files, database files, and build artifacts while preserving directory structure.

## Files Updated

### Root `.gitignore`

Added patterns for:

- **Binaries**: `bin/`, `*.exe`, `*.dll`, `*.so`, `*.dylib`
- **Go build artifacts**: `apps/server/main`, `apps/server/main.exe`
- **Database files**: `*.db`, `*.db-shm`, `*.db-wal`
- **Data directory**: Ignores contents but preserves structure

### Server `.gitignore` (`apps/server/.gitignore`)

Server-specific patterns for:

- Go binaries and test artifacts
- Database files in data directory
- Configuration files (keeps examples)
- Temporary files

## Directory Structure Preservation

### `.gitkeep` Files Created

1. `apps/server/data/.gitkeep` - Preserves data directory
2. `apps/server/data/media/.gitkeep` - Preserves media subdirectory
3. `bin/.gitkeep` - Preserves bin directory

These files ensure empty directories are tracked by Git while their contents are ignored.

## Ignored Files

### Binaries

- `bin/*.exe` - Compiled Go binaries
- `apps/server/main` - Go server binary (Linux/Mac)
- `apps/server/main.exe` - Go server binary (Windows)
- `*.dll`, `*.so`, `*.dylib` - Shared libraries

### Database Files

- `apps/server/data/whatsmeow.db` - WhatsApp protocol database
- `apps/server/data/application.db` - Application database
- `*.db-shm`, `*.db-wal` - SQLite temporary files

### Media Files

- `apps/server/data/media/*` - All uploaded media files

### Configuration

- `apps/server/config.yaml` - Local configuration
- Keeps: `config.example.yaml`, `config.example.json`

## Verification

To check if files are properly ignored:

```bash
# Check specific files
git check-ignore -v apps/server/data/*.db
git check-ignore -v apps/server/main.exe
git check-ignore -v bin/*.exe

# List all ignored files in a directory
git status --ignored apps/server/data/
```

## Best Practices

1. **Never commit database files** - They contain local data and can be large
2. **Never commit binaries** - They should be built from source
3. **Keep example configs** - Help other developers set up their environment
4. **Preserve directory structure** - Use .gitkeep for empty directories
5. **Document ignored patterns** - Make it clear what's excluded and why

## Migration Notes

If you previously committed database or binary files:

```bash
# Remove from Git but keep locally
git rm --cached apps/server/data/*.db
git rm --cached apps/server/main.exe
git rm --cached bin/*.exe

# Commit the removal
git commit -m "Remove database and binary files from Git"
```
