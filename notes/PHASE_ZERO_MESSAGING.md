# Phase 0 — Plain-text WebSocket Messaging

Goal: verify that the full message delivery pipeline (REST + WebSocket + MongoDB)
is functional before adding any cryptography. Messages are stored and transmitted
as plain text. Everything here is a foundation that Phase 1 (E2EE) will build on
without structural changes.

---

## New dependency

`github.com/gorilla/websocket` — the standard choice for WebSocket in Go/Gin.
No other new dependencies. MongoDB driver is already wired (`config.MongoDB`).

---

## File layout

```
internal/modules/messaging/
  entity.go                    ← Conversation + Message mongo document types
  repository.go                ← MongoDB CRUD (SaveMessage, GetThread,
                                  GetOrCreateConversation)
  hub.go                       ← in-memory connection registry (Hub + Client)
  connect/
    controller.go              ← GET  /api/messaging/ws  (WS upgrade)
  conversations/
    controller.go              ← POST /api/messaging/conversations
    dto.go
    service.go
  send/
    controller.go              ← POST /api/messaging/messages
    dto.go
    service.go
  get_thread/
    controller.go              ← GET  /api/messaging/conversations/:id/messages
    dto.go
    service.go

internal/http/routes/
  messaging_routes.go          ← RegisterMessagingRoutes

config/mongodb.go              ← already done
internal/container/container.go ← add MongoDB *mongo.Database field
cmd/main.go                    ← call config.InitMongoDB(), pass to container
```

---

## MongoDB collections

### `conversations`
```json
{
  "_id":              ObjectID,
  "participant_ids":  [uint64, uint64],
  "created_at":       ISODate,
  "last_message_at":  ISODate
}
```
Index: `{ participant_ids: 1 }` (for inbox lookup by user ID).

### `messages`
```json
{
  "_id":              ObjectID,
  "conversation_id":  ObjectID,
  "sender_id":        uint64,
  "content":          "string (plain text in Phase 0)",
  "type":             "text",
  "created_at":       ISODate,
  "delivery_status":  "sent | delivered | read"
}
```
Index: `{ conversation_id: 1, created_at: -1 }` (paginated thread fetch).

---

## Go entities (`entity.go`)

```go
type Conversation struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"`
    ParticipantIDs []uint64           `bson:"participant_ids"`
    CreatedAt      time.Time          `bson:"created_at"`
    LastMessageAt  time.Time          `bson:"last_message_at"`
}

type Message struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"`
    ConversationID primitive.ObjectID `bson:"conversation_id"`
    SenderID       uint64             `bson:"sender_id"`
    Content        string             `bson:"content"`
    Type           string             `bson:"type"` // "text" for Phase 0
    CreatedAt      time.Time          `bson:"created_at"`
    Status         string             `bson:"delivery_status"`
}
```

---

## WebSocket hub (`hub.go`)

The Hub is a single long-lived goroutine that serialises all connection state
changes through channels — avoids mutexes on the connection map.

```go
type Client struct {
    UserID uint64
    send   chan []byte       // outbound messages queued here
    conn   *websocket.Conn
}

type Hub struct {
    clients    map[uint64]*Client
    register   chan *Client
    unregister chan *Client
    deliver    chan Delivery  // {RecipientID, payload}
}
```

Hub.Run() loop:
- `register`   → add to map, close previous connection for same user ID if any
- `unregister` → delete from map, close send channel
- `deliver`    → look up recipient; if present, non-blocking send to client.send

Each Client runs two goroutines:
- **readPump**: reads from `conn` (handles close/ping); sends unregister on EOF
- **writePump**: reads from `send` channel, writes to `conn`; handles ping/pong

---

## WebSocket endpoint — `connect/`

`GET /api/messaging/ws?token=<JWT>`

JWT is taken from the `token` query param because the browser WebSocket API
does not support custom headers during the handshake. The handler validates
the token, creates a `Client`, registers it with the Hub, then starts read/
writePump goroutines. Gin's handler returns immediately; the goroutines keep
the connection alive.

```
GET /api/messaging/ws?token=eyJ...
→ upgrade → register Client{userID=42} in Hub
→ readPump / writePump goroutines started
→ Gin handler returns (connection lives in goroutines)
```

---

## Conversations endpoint — `conversations/`

`POST /api/messaging/conversations`
```json
Body: { "other_user_id": 7 }
```
Service: query conversations where `participant_ids` contains BOTH the caller's ID
and `other_user_id`. If found, return it. If not, create it.
Returns the conversation ID (used as a channel key for all subsequent send/get).

This ensures only one conversation document exists per pair, regardless of how
many times the endpoint is called.

---

## Send endpoint — `send/`

`POST /api/messaging/messages`
```json
Body: {
  "conversation_id": "6871abc...",
  "content":         "Hello!"
}
```

Service steps:
1. Validate the caller is a participant in the conversation (load from Mongo).
2. Insert `Message` doc into `messages` with `status = "sent"`.
3. Determine recipient ID (the participant that is NOT the caller).
4. Try to deliver via Hub (`hub.deliver <- Delivery{RecipientID, payload}`).
   - Hub attempts non-blocking write; if recipient is offline, message stays
     in Mongo with status "sent" — client fetches history on next open.
   - If recipient is online, Hub writes to their send channel; service
     updates status to "delivered" in Mongo.
5. Return 201 with the saved message.

The JSON payload pushed over WebSocket to the recipient:
```json
{
  "id":              "6871abc...",
  "conversation_id": "6871abc...",
  "sender_id":       42,
  "content":         "Hello!",
  "type":            "text",
  "created_at":      "2026-06-24T12:00:00Z",
  "status":          "delivered"
}
```

---

## Get thread endpoint — `get_thread/`

`GET /api/messaging/conversations/:id/messages?before=<ISO8601>&limit=<n>`

Paginated by cursor (`before` = created_at of the oldest message the client
already has; defaults to now). Returns up to `limit` messages (default 50,
max 100) in descending order, which the client reverses for display.

Service validates the caller is a participant before querying.

---

## Wiring into the app

### `AppContainer`
```go
type AppContainer struct {
    // existing fields ...
    MongoDB *mongo.Database
    Hub     *messaging.Hub   // shared singleton
}
```

### `cmd/main.go`
```go
config.InitMongoDB()
hub := messaging.NewHub()
go hub.Run()

appContainer := container.NewAppContainer(
    // existing args ...
    config.MongoDB,
    hub,
)
```

### `router.go`
```go
routes.RegisterMessagingRoutes(router, appContainer)
```

### `messaging_routes.go`
```go
// All routes require JWT (token query param for WS, header for REST)
g := router.Group("/api/messaging")
g.Use(middleware.JWT(appContainer.JWTService))

connect.RegisterRoutes(g, appContainer.Hub, appContainer.JWTService)
conversations.RegisterRoutes(g, appContainer.MongoDB)
send.RegisterRoutes(g, appContainer.MongoDB, appContainer.Hub)
get_thread.RegisterRoutes(g, appContainer.MongoDB)
```

Note: the WS endpoint reads the JWT from the query param itself (before the
upgrade), so the standard JWT middleware can be skipped on that specific route
and validation done inline in the connect controller.

---

## What Phase 1 (E2EE) changes

Nothing structural changes. The swap is:
- `Message.Content` (plain string) → `Message.Ciphertext` (base64 blob) + `Message.Nonce`
- Client encrypts before `POST /messages`; recipient decrypts after receiving
  the WS push or fetching history
- Server code is untouched — it stores and relays opaque bytes instead of
  readable strings

The Hub, Repository, routes, and MongoDB schema all carry forward unchanged.

---

## Out of scope for Phase 0

- Read receipts / "delivered" status update from recipient client
- Typing indicators
- Presence (online/offline)
- Pagination on conversation list (inbox)
- Any crypto
