# WebSockets — How They Work and Why They Fit Real-Time Messaging

---

## The problem with plain HTTP

Every HTTP request follows a strict request-response cycle:

```
Client ──── GET /messages ────► Server
Client ◄─── 200 OK + data ───── Server
(connection closes)
```

The client always has to ask first. The server cannot spontaneously push data.
For a chat app this means the client would have to keep asking "are there new
messages?" on a timer — a technique called **polling**. Polling is wasteful:
most requests come back empty, and there is always a delay between a message
being sent and the recipient seeing it (however short your poll interval is).

Two improvements exist before reaching WebSockets:

**Long polling** — the client sends a request, the server holds it open until
a new message arrives, then responds and the client immediately opens another
request. Better than polling, but still half-duplex (one direction at a time)
and each round trip carries full HTTP overhead (headers, TLS handshake, etc.)

**Server-Sent Events (SSE)** — a persistent one-way stream from server to
client over HTTP. Good for dashboards and notifications, but the client cannot
send data back over the same connection — it still needs separate POST requests.

---

## What WebSockets are

WebSocket is a protocol (RFC 6455) that upgrades an existing HTTP connection
into a persistent, **full-duplex** (bidirectional, simultaneously) channel.
After the upgrade both sides can send frames to the other at any time without
waiting for a request.

```
Client ◄──────────────────────► Server
         persistent TCP channel
         both sides can write
         at any moment
```

One TCP connection stays open for the entire session. Frames are lightweight —
they carry a 2-byte minimum header instead of the kilobytes of HTTP headers on
every exchange.

---

## The handshake

WebSocket starts as a normal HTTP request and upgrades:

```
Client → Server  (HTTP request)
GET /api/messaging/ws HTTP/1.1
Host: example.com
Connection: Upgrade
Upgrade: websocket
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13

Server → Client  (HTTP response)
HTTP/1.1 101 Switching Protocols
Connection: Upgrade
Upgrade: websocket
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

`101 Switching Protocols` means: "agreed, this TCP connection is no longer
HTTP — both of us now speak the WebSocket frame protocol on it."

After this single exchange the HTTP layer is gone. What remains is a raw
TCP connection with a thin framing layer on top. Both sides can now write
whenever they have something to say.

---

## Frame structure (simplified)

Every WebSocket message is wrapped in a frame:

```
 FIN  opcode  MASK  payload length  [masking key]  payload
  1     4      1       7 / 16 / 64      0 or 4       N bytes
```

- **opcode** distinguishes text frames, binary frames, ping, pong, and close.
- **FIN** marks the last fragment of a multi-frame message.
- **MASK** — frames sent from client to server must be masked (XOR with a
  4-byte key included in the frame). Server-to-client frames are unmasked.
  This is a spec requirement, not encryption.

The frame overhead for a small text message is 2–6 bytes vs. 200–800 bytes of
HTTP headers for the equivalent HTTP request. For thousands of messages per
second this difference is significant.

---

## Connection lifecycle

```
1. Client connects (HTTP upgrade handshake above)
2. Both sides exchange frames freely
3. Either side sends a Close frame when done
4. The other side acknowledges with a Close frame
5. TCP connection is torn down
```

Servers typically run a **ping / pong** heartbeat (WebSocket has built-in ping
and pong frames) to detect half-open connections — TCP does not reliably
surface a dead peer immediately, so periodic pings let both sides know the
connection is still alive.

---

## Why WebSockets are the right choice for this messaging system

| Requirement | Why WebSocket fits |
|-------------|-------------------|
| **Instant delivery** | Server pushes to the recipient the moment a message is saved — no polling delay |
| **Low overhead at scale** | One persistent connection per user instead of a new HTTP round trip per message; no repeated headers |
| **Bidirectional** | Client can send and receive on the same connection — natural fit for a chat channel |
| **Presence and status** | Connection open = user is online; disconnect event fires immediately on drop, enabling reliable online/offline tracking without polling |
| **Foundation for typing indicators** | Lightweight frames make it cheap to send "user is typing" signals that would be absurd to poll for |
| **Future E2EE** | The channel is transport-agnostic — the payload is an opaque blob whether it contains plaintext or ciphertext. Switching to encrypted messages requires no changes to the WebSocket layer |

SSE would cover server-push but the client would still need HTTP POST to send
messages, complicating the delivery-status update flow (you need to know the
message arrived before you can mark it "delivered"). WebSocket keeps both
directions on one connection, which makes the send → store → push → confirm
cycle straightforward to implement.

---

## How it maps to this project

```
Angular client                   Go server (Gin + gorilla/websocket)
──────────────                   ──────────────────────────────────
new WebSocket(url)          →    GET /api/messaging/ws?token=<JWT>
                            ←    101 Switching Protocols
                                 Hub registers Client{userID}

POST /api/messaging/messages →   save to MongoDB
                            ←    HTTP 201 (message saved)
                                 Hub.deliver → recipient's send channel
                            →    WS frame pushed to recipient's connection
recipient receives frame    ←    (no polling, no delay)
```

The Hub is the in-process router: it holds a `map[userID → Client]` and,
when a message is saved, looks up the recipient and writes the payload to
their outbound channel. If the recipient is offline, the message sits in
MongoDB and is fetched via `GET /conversations/:id/messages` when they
reconnect — the two delivery paths (push vs. pull) complement each other.
