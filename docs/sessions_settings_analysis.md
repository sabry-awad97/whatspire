# Sessions & Settings Integration Analysis

> **Focus**: Deep analysis of `/sessions` and `/settings` routes for full functionality enhancement

## Executive Summary

The **Sessions** module is **well-integrated** with comprehensive hook usage. The **Settings** module has a **mixed integration state** - API Keys are fully functional, but General settings (API endpoint/key) and Webhook configuration lack backend persistence.

---

## Sessions Route Analysis

### Route Structure

| Route | File | Integration Status |
|-------|------|-------------------|
| `/sessions` | [index.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/sessions/index.tsx) | ✅ **Fully Integrated** |
| `/sessions/new` | [new.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/sessions/new.tsx) | ✅ **Fully Integrated** |
| `/sessions/:id` | [$sessionId.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/sessions/$sessionId.tsx) | ✅ **Fully Integrated** |
| `/sessions/:id/edit` | [edit.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/sessions/$sessionId/edit.tsx) | ⚠️ **Partial - TODO** |
| `/sessions/:id/webhooks` | [webhooks.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/sessions/$sessionId/webhooks.tsx) | ❌ **Not Integrated** |

---

### Session Components Deep Dive

#### 1. SessionDetails ([session-details.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/components/sessions/session-details.tsx))
**Lines**: 1,500 | **Status**: ✅ **Excellent Integration**

**Hooks Used**:
| Hook | Purpose | Status |
|------|---------|--------|
| `useApiClient` | API client context | ✅ Active |
| `useContacts` | Fetch session contacts | ✅ Active (lines 96-103) |
| `useEvents` | Fetch session events | ✅ Active (lines 105-116) |
| `useSendMessage` | Send WhatsApp messages | ✅ Active (lines 119-128) |
| `useReconnectSession` | Reconnect session | ✅ Active (lines 130-137) |
| `useDisconnectSession` | Disconnect session | ✅ Active (lines 139-146) |
| `useDeleteSession` | Delete session | ✅ Active (lines 148-156) |

**Features**:
- ✅ QR code display/refresh via WebSocket
- ✅ Session status display with real-time updates
- ✅ Test message sending (GUI + code snippets for 5 languages)
- ✅ API credentials tab with token management
- ✅ Recent Activity tab showing message events
- ✅ Session Logs tab showing status changes
- ✅ Webhook navigation button

**Minor Gap**:
- API token read from `localStorage` (line 89) - works but relies on Settings tab to populate it

---

#### 2. SessionCard ([session-card.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/components/sessions/session-card.tsx))
**Lines**: 324 | **Status**: ✅ **Fully Integrated**

**Hooks Used**:
- `useApiClient`, `useReconnectSession`, `useDisconnectSession`, `useDeleteSession`

**Features**:
- ✅ Status display with icons (connected/pending/disconnected/connecting/logged_out)
- ✅ Reconnect/Disconnect actions via hooks
- ✅ Delete with confirmation dialog
- ✅ Edit/Manage navigation buttons
- ✅ Loading states during mutations

---

#### 3. QRCodeDisplay ([qr-code-display.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/components/sessions/qr-code-display.tsx))
**Lines**: 153 | **Status**: ✅ **Fully Integrated**

**Hooks Used**:
- `useQRWebSocket` - Real-time QR code updates via WebSocket

**Features**:
- ✅ WebSocket-based QR code streaming
- ✅ Status states: loading → ready → authenticated/expired
- ✅ Refresh capability to reconnect WebSocket
- ✅ Toast notifications for authentication events

---

#### 4. Edit Session Route ([edit.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/sessions/$sessionId/edit.tsx))
**Lines**: 537 | **Status**: ⚠️ **Partial Integration - Critical Gap**

**Issue Identified at Line 57**:
```typescript
onSubmit: async ({ value }) => {
  try {
    // TODO: Implement update session API endpoint  ← CRITICAL GAP
    toast.success("Session updated successfully");
    navigate({ to: "/sessions/$sessionId", ... });
  } catch (error) {
    toast.error("Failed to update session");
  }
}
```

**Current Behavior**:
- Form collects all session settings (name, phone, protection, logging, filtering, webhook)
- Form submit shows success toast but **does NOT call any backend API**
- Uses `useSession` hook to fetch session data (working)
- Uses `@tanstack/react-form` for form management (working)

**Fix Required**:
- Create/use `useUpdateSession` hook
- Backend endpoint: `PATCH /api/sessions/:id` or `PUT /api/sessions/:id`

---

#### 5. Webhooks Configuration Route ([webhooks.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/sessions/$sessionId/webhooks.tsx))
**Lines**: 449 | **Status**: ❌ **Not Integrated - UI Only**

**Hooks Used**:
- `useSession` - To fetch session data (working)

**Critical Gaps**:

| Function | Line | Current Implementation | Required |
|----------|------|----------------------|----------|
| `handleSave` | 127-129 | `toast.success("Webhook configuration saved")` | API call to save webhook |
| `handleRotateSecret` | 135-137 | `toast.success("Webhook secret rotated")` | API call to rotate secret |

**UI Features (Frontend-Only)**:
- ✅ Endpoint toggle (enabled/disabled)
- ✅ Webhook URL input
- ✅ Webhook secret display with show/hide
- ✅ 21 event type subscriptions
- ✅ Message filtering options (ignore groups/broadcasts/channels)

**Fix Required**:
- Create `useWebhookConfig` and `useUpdateWebhookConfig` hooks
- Backend endpoints:
  - `GET /api/sessions/:id/webhook` - Fetch current config
  - `PUT /api/sessions/:id/webhook` - Update config
  - `POST /api/sessions/:id/webhook/rotate-secret` - Rotate secret

---

## Settings Route Analysis

### Route Structure

| Route | File | Integration Status |
|-------|------|-------------------|
| `/settings` | [index.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/routes/settings/index.tsx) | Mixed |

---

### Settings Components Deep Dive

#### 1. Settings Main ([settings.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/components/settings/settings.tsx))
**Lines**: 172 | **Status**: ❌ **General Tab Not Integrated**

**Tabs Analysis**:

| Tab | Status | Notes |
|-----|--------|-------|
| General | ❌ **Not Integrated** | API settings use `useState` only |
| API Keys | ✅ **Fully Integrated** | Uses `APIKeySettings` component |

**General Tab Critical Issues**:

```typescript
// Lines 22-25: Local state only, not persisted
const [apiEndpoint, setApiEndpoint] = useState("https://api.whatspire.com");
const [apiKey, setApiKey] = useState("");

// Line 36-38: Shows toast but doesn't save
const handleSaveApiSettings = () => {
  toast.success("API settings saved successfully");
};

// Lines 44-49: Simulated connection test
const handleTestConnection = async () => {
  toast.info("Testing connection...");
  await new Promise((resolve) => setTimeout(resolve, 1500)); // Fake delay
  toast.success("Connection successful!"); // Always succeeds
};
```

**Current Behavior**:
- API endpoint defaults to `https://api.whatspire.com`
- API key input field exists but value is lost on refresh
- "Test Connection" button shows fake success after 1.5s delay
- "Save API Settings" button shows success toast but doesn't persist

---

#### 2. APIKeySettings ([api-key-settings.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/components/settings/api-key-settings.tsx))
**Lines**: 127 | **Status**: ✅ **Fully Integrated**

**Features**:
- ✅ Header with Create API Key button
- ✅ Info banner about API keys
- ✅ Pagination, filtering, refresh
- ✅ Delegates to `APIKeyList` and `CreateAPIKeyDialog`

---

#### 3. APIKeyList ([api-key-list.tsx](file:///d:/programming/better-desktop-apps/whatspire/apps/web/src/components/settings/api-key-list.tsx))
**Lines**: 629 | **Status**: ✅ **Fully Integrated**

**Hooks Used**:
- `useAPIKeys(apiClient, { page, limit, ...filters })` - Paginated API key listing

**Features**:
- ✅ Role filter (read/write/admin)
- ✅ Status filter (active/revoked)
- ✅ Pagination with page size control (10/25/50/100)
- ✅ Key visibility toggle with auto-hide after 10s
- ✅ Copy to clipboard
- ✅ Revoke dialog integration
- ✅ Details dialog integration
- ✅ Empty, loading, and error states

---

## Available Hooks Inventory

Based on the `@whatspire/hooks` package:

### Session Hooks
| Hook | File | Purpose |
|------|------|---------|
| `useSessions` | use-sessions.ts:37 | List all sessions |
| `useSession` | use-sessions.ts:68 | Get single session |
| `useCreateSession` | use-sessions.ts:107 | Create new session |
| `useDeleteSession` | use-sessions.ts:140 | Delete session |
| `useReconnectSession` | use-sessions.ts:173 | Reconnect session |
| `useDisconnectSession` | use-sessions.ts:206 | Disconnect session |

### Message Hooks
| Hook | File | Purpose |
|------|------|---------|
| `useSendMessage` | use-messages.ts:30 | Send message |
| `useSendPresence` | use-messages.ts:50 | Send presence |
| `useSendReaction` | use-messages.ts:70 | Send reaction |
| `useRemoveReaction` | use-messages.ts:90 | Remove reaction |
| `useSendReceipt` | use-messages.ts:110 | Send read receipt |

### Data Hooks
| Hook | File | Purpose |
|------|------|---------|
| `useContacts` | use-contacts.ts:25 | List contacts |
| `useContactProfile` | use-contacts.ts:49 | Get contact profile |
| `useChats` | use-contacts.ts:73 | List chats |
| `useCheckPhone` | use-contacts.ts:97 | Check phone registration |
| `useGroups` | use-groups.ts:21 | List groups |
| `useEvents` | use-events.ts:21 | List events |
| `useAPIKeys` | use-api-keys.ts:36 | List API keys |
| `useCreateAPIKey` | use-api-keys.ts:70 | Create API key |

### WebSocket Hooks
| Hook | File | Purpose |
|------|------|---------|
| `useWebSocket` | use-websocket.ts:128 | Generic WebSocket |
| `useQRWebSocket` | use-websocket.ts:302 | QR code WebSocket |
| `useEventWebSocket` | use-websocket.ts:340 | Event WebSocket |

---

## Missing Hooks (To Be Created)

| Hook | Purpose | Used By |
|------|---------|---------|
| `useUpdateSession` | Update session settings | edit.tsx |
| `useWebhookConfig` | Get webhook configuration | webhooks.tsx |
| `useUpdateWebhookConfig` | Update webhook configuration | webhooks.tsx |
| `useRotateWebhookSecret` | Rotate webhook secret | webhooks.tsx |
| `useAppSettings` | Get/set app settings | settings.tsx |
| `useTestConnection` | Test API connection | settings.tsx |

---

## Priority Gap Summary

### 🔴 Critical (Blocking Functionality)

1. **edit.tsx - Session Update** (Line 57)
   - Form collects data but doesn't save
   - Need `useUpdateSession` hook
   - Backend: `PATCH /api/sessions/:id`

2. **settings.tsx - API Configuration**
   - API endpoint/key not persisted
   - Values lost on refresh
   - Session details relies on localStorage token

### 🟠 High (Major Missing Feature)

3. **webhooks.tsx - Webhook Configuration**
   - Full UI exists but no backend integration
   - Need 3 new hooks and backend endpoints

### 🟡 Medium (UX Improvement)

4. **settings.tsx - Connection Test**
   - Currently fakes success
   - Should call actual health endpoint

---

## Recommended Implementation Order

### Phase 1: Session Edit Fix (Est. 2 hours)
1. Create `useUpdateSession` hook
2. Add backend endpoint if missing
3. Update edit.tsx to call the hook

### Phase 2: Settings Persistence (Est. 2 hours)
1. Implement localStorage or backend persistence for API config
2. Sync with ApiClient context
3. Real connection test using health endpoint

### Phase 3: Webhook Integration (Est. 4 hours)
1. Create webhook hooks
2. Add backend endpoints if missing
3. Wire up webhooks.tsx to use hooks
4. Load existing config on page load

---

## Technical Details

### Current localStorage Usage

| Key | Used By | Purpose |
|-----|---------|---------|
| `whatspire_api_token` | session-details.tsx:89 | Store active API key for sending requests |

### API Client Context

The `useApiClient` hook provides the API client from context, which uses configuration from environment variables:
- `VITE_API_BASE_URL` - Backend API URL (default: `http://localhost:8080`)

### Real-Time Updates

Session status updates are handled via:
- `useQRWebSocket` - For QR code streaming during authentication
- `useEventWebSocket` - For session event notifications
- `useSessionEvents` - Context-based event aggregation
