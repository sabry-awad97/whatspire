# ADR-004: React + Tauri for Desktop Application

**Date**: 2026-02-03  
**Status**: Accepted  
**Deciders**: Development Team  
**Technical Story**: Desktop User Experience - Cross-Platform Desktop App

---

## Context and Problem Statement

End users needed a desktop application to manage WhatsApp sessions with a modern, responsive UI. The application required cross-platform support (Windows, macOS, Linux), real-time updates, and a native feel. Which technology stack should we use for the desktop application?

## Decision Drivers

- Cross-platform support (Windows, macOS, Linux)
- Modern, responsive UI with glassmorphic design
- Real-time updates via WebSocket
- Small bundle size and fast startup
- Native system integration (notifications, tray icon)
- Developer productivity and maintainability
- Type safety and good tooling
- Active community and ecosystem

## Considered Options

- **Option 1**: React + Tauri (Rust-based)
- **Option 2**: Electron (Chromium-based)
- **Option 3**: React Native Desktop
- **Option 4**: Native apps (Swift, Kotlin, C++)

## Decision Outcome

Chosen option: "**React + Tauri**", because it provides the best balance of performance, bundle size, and developer experience while delivering a truly native desktop application.

### Positive Consequences

- **Small Bundle Size**: ~10MB vs 100MB+ for Electron
- **Fast Startup**: Native Rust backend, no Chromium overhead
- **Native Feel**: Uses system WebView, feels like native app
- **Security**: Rust's memory safety, sandboxed architecture
- **Performance**: Lower memory usage than Electron
- **Modern UI**: React ecosystem for UI components
- **Type Safety**: TypeScript for frontend, Rust for backend
- **Cross-Platform**: Single codebase for all platforms

### Negative Consequences

- **Newer Technology**: Smaller community than Electron
- **Learning Curve**: Team needs to learn Tauri concepts
- **WebView Differences**: Slight rendering differences across platforms
- **Limited Native APIs**: Some platform-specific features require custom Rust code
- **Debugging**: More complex than pure web debugging

## Pros and Cons of the Options

### Option 1: React + Tauri

- Good, because 10x smaller bundle size than Electron
- Good, because faster startup and lower memory usage
- Good, because uses system WebView (native feel)
- Good, because Rust backend for security and performance
- Good, because React ecosystem for UI
- Bad, because newer technology with smaller community
- Bad, because WebView differences across platforms
- Bad, because requires Rust knowledge for native features

### Option 2: Electron

- Good, because mature and battle-tested
- Good, because large community and ecosystem
- Good, because consistent rendering across platforms
- Good, because extensive native API support
- Bad, because 100MB+ bundle size
- Bad, because high memory usage (full Chromium)
- Bad, because slower startup time
- Bad, because security concerns (Node.js in renderer)

### Option 3: React Native Desktop

- Good, because React Native experience
- Good, because native components
- Bad, because experimental and unstable
- Bad, because limited desktop support
- Bad, because smaller ecosystem than web
- Bad, because performance issues on desktop

### Option 4: Native Apps

- Good, because best performance
- Good, because full platform integration
- Good, because native look and feel
- Bad, because 3 separate codebases (Swift, Kotlin, C++)
- Bad, because 3x development time
- Bad, because harder to maintain consistency
- Bad, because requires platform-specific expertise

## Technology Stack

### Frontend

- **React 19.2.4**: UI framework
- **TanStack Router**: File-based routing
- **TanStack Query**: Data fetching and caching
- **TanStack Form**: Form management
- **Zustand**: State management
- **Tailwind CSS 4**: Styling with OKLCH colors
- **shadcn UI**: Component library
- **TypeScript**: Type safety

### Backend (Tauri)

- **Tauri 2.9.6**: Desktop framework
- **Rust**: Native backend
- **System WebView**: Rendering engine

### Build Tools

- **Vite 7.3.1**: Fast build tool
- **Bun**: Package manager and runtime

## Architecture

```
apps/web/
├── src/                    # React application
│   ├── components/        # UI components
│   ├── routes/            # TanStack Router routes
│   ├── lib/               # API client, WebSocket
│   └── stores/            # Zustand stores
├── src-tauri/             # Tauri backend
│   ├── src/
│   │   ├── main.rs       # Rust entry point
│   │   └── lib.rs        # Tauri commands
│   ├── capabilities/      # Permission system
│   └── icons/             # App icons
└── vite.config.ts         # Vite configuration
```

## Design System

### Glassmorphic Theme

- **Color Space**: OKLCH for perceptual uniformity
- **Background**: `oklch(0.12 0.02 250)` - Deep Night
- **Primary**: `oklch(0.70 0.15 180)` - Teal
- **Accent**: `oklch(0.80 0.15 90)` - Amber
- **Effects**: Backdrop blur, gradient borders, glow effects

### Components

- **Sessions**: Manage WhatsApp sessions with QR codes
- **Messages**: View and filter messages
- **Contacts**: Manage contacts with avatars
- **Groups**: View and manage groups
- **Settings**: Configure API, theme, notifications

## Real-Time Updates

### WebSocket Integration

```typescript
// WebSocket manager for real-time updates
const ws = new WebSocketManager("ws://localhost:8080/ws");

ws.on("qr_code", (data) => {
  // Update QR code display
});

ws.on("message", (data) => {
  // Add new message to list
});

ws.on("session_status", (data) => {
  // Update session status
});
```

## Platform-Specific Considerations

### Windows

- Uses Edge WebView2
- Requires WebView2 runtime (auto-installed)
- Native window controls
- System tray integration

### macOS

- Uses WKWebView
- Native title bar
- Menu bar integration
- Notification center

### Linux

- Uses WebKitGTK
- Multiple desktop environments supported
- System tray (where available)
- Native notifications

## Build and Distribution

### Development

```bash
bun run desktop:dev    # Start dev server with Tauri
```

### Production Build

```bash
bun run desktop:build  # Build for current platform
```

### Distribution

- **Windows**: MSI installer, portable EXE
- **macOS**: DMG, app bundle
- **Linux**: AppImage, deb, rpm

## Links

- [Tauri Documentation](https://tauri.app/)
- [React Documentation](https://react.dev/)
- [TanStack Documentation](https://tanstack.com/)
- Related: ADR-001 (Clean Architecture)
- See: `apps/web/PHASE7_SUMMARY.md` for implementation details

---

## Notes

### Bundle Size Comparison

- **Tauri**: ~10MB (uses system WebView)
- **Electron**: ~100MB (includes Chromium)
- **Native**: ~5MB (platform-specific)

### Memory Usage

- **Tauri**: ~50MB base + app memory
- **Electron**: ~150MB base + app memory
- **Native**: ~20MB base + app memory

### Startup Time

- **Tauri**: <1 second
- **Electron**: 2-3 seconds
- **Native**: <0.5 seconds

### Security Model

Tauri uses a capability-based security model:

- No Node.js in renderer (unlike Electron)
- Explicit permission system
- Sandboxed WebView
- Rust backend for sensitive operations

### Auto-Updates

Tauri supports auto-updates:

- Check for updates on startup
- Download in background
- Install on next launch
- Rollback on failure

### Testing Strategy

- **Unit Tests**: Vitest for React components
- **Integration Tests**: Test Tauri commands
- **E2E Tests**: Playwright for full app testing
- **Manual Testing**: Test on all platforms

### Known Limitations

- **WebView Differences**: Slight rendering differences across platforms
  - Solution: Test on all platforms, use feature detection
- **Native APIs**: Some features require custom Rust code
  - Solution: Use Tauri plugins or write custom commands
- **Debugging**: More complex than pure web
  - Solution: Use Tauri DevTools, Rust debugging tools

### Migration from Electron

If migrating from Electron:

1. Replace IPC with Tauri commands
2. Remove Node.js dependencies from renderer
3. Rewrite native modules in Rust
4. Update build configuration
5. Test on all platforms

### Future Enhancements

- **Mobile Support**: Tauri Mobile (iOS, Android)
- **Plugins**: Custom Tauri plugins for native features
- **Auto-Updates**: Implement update mechanism
- **Crash Reporting**: Add Sentry or similar
- **Analytics**: Add usage analytics
