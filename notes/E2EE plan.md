# Plan: End-to-End Encrypted Messaging (User ↔ Craftsman) with MongoDB

## Goal

Add direct messaging between customers and craftsmen (e.g. negotiating a
custom order) on top of the existing Go/Gin/PostgreSQL backend, with:
- End-to-end encryption (server never sees plaintext content)
- Support for text, attachments, "real-time" photos/videos, and voice
  recordings
- MongoDB as the store for message documents/attachments metadata, separate
  from the existing relational (PostgreSQL/GORM) data

This is additive — it does not replace the existing Postgres schema (users,
orders, products, etc.); it's a new bounded subsystem.

## Why MongoDB here

Messages are naturally document-shaped (variable attachment arrays, mixed
content types, growing thread history) and you'll want to query by
conversation/timestamp range rather than relational joins — a good fit for
Mongo over adding more Postgres tables. Keep referential fields (sender
user_id, receiver user_id/craftsman_id) as plain IDs pointing back into
Postgres; Mongo does not need to know about Postgres schema, just store the
IDs as opaque references.

## High-level architecture

```
Client A (Angular)  <---- WebSocket / HTTPS ---->  Go API  <---->  MongoDB
Client B (Angular)  <---- WebSocket / HTTPS ---->  (relay only,
                                                     never decrypts)
```

The Go backend is a **relay and metadata store**, not a participant in the
encryption. It never has access to plaintext or to the symmetric keys used
to encrypt message bodies/media.

## E2EE design

### Key management

- Each user/craftsman device generates an asymmetric identity key pair
  client-side (e.g. X25519 for key agreement, Ed25519 for signing) on first
  use of messaging — use a vetted library, don't hand-roll crypto
  (e.g. `libsodium`/`tweetnacl` via WebCrypto wrapper, or adopt the
  Signal protocol's Double Ratchet via an existing JS implementation such
  as `libsignal-protocol-javascript` or a maintained fork).
- Public keys are uploaded to the backend (`POST /messaging/keys`) and
  stored in Mongo (or Postgres, since it's small and relational-ish) keyed
  by user_id + device_id. Private keys never leave the client (stored in
  browser `IndexedDB`, not localStorage, since IndexedDB supports
  non-extractable CryptoKey objects via WebCrypto).
- On starting a conversation, clients perform a key exchange (X3DH-style:
  fetch the other party's public identity key + one-time prekey from the
  server, derive a shared secret) to establish a session key, then use a
  ratcheting scheme (Double Ratchet) so each message uses a fresh key and
  past messages stay safe even if a later key leaks (forward secrecy).
- Recommend **not implementing the ratchet from scratch**. Evaluate using
  an existing audited library/protocol implementation (Signal's protocol
  libraries, or Matrix's Olm/Megolm if you want group-chat-ready crypto)
  rather than writing custom AES/X25519 glue — this is the single highest
  -risk part of the project to get wrong.

### Message encryption

- Text messages: encrypt plaintext client-side with the session's derived
  AES-256-GCM key before sending; ciphertext + nonce/IV + auth tag is what
  goes over the wire and what gets stored in Mongo.
- Attachments (photos/videos/voice recordings):
  1. Client encrypts the file blob client-side (AES-256-GCM, streaming for
     large files / chunked encryption for video to avoid loading whole file
     in memory).
  2. Encrypted blob is uploaded to backend storage (reuse the existing
     `files` module's `Storage` interface — it's already an abstraction
     over local filesystem, could swap to S3-compatible storage later
     without changing callers).
  3. Mongo stores only: encrypted blob URL/reference, content type, size,
     encrypted symmetric key (itself encrypted with the recipient's public
     key — "envelope encryption"), nonce, checksum of ciphertext (integrity
     check, not a decryption mechanism).
  4. Recipient downloads encrypted blob, decrypts the envelope key with
     their private key, then decrypts the blob locally.
- "Real-time" photos/video/voice (i.e. captured live in-app, not picked from
  gallery): same pipeline — capture via `MediaRecorder`/`getUserMedia` in
  the browser, encrypt the resulting blob client-side before upload. No
  different from a regular attachment from the architecture's point of
  view; the only frontend-specific piece is capture UI, not the crypto.

### What the server CAN see (by design)

- Sender/receiver IDs, timestamps, conversation IDs, message size,
  content-type, delivery/read status — needed for inbox UX, push
  notifications, and abuse/rate-limiting. This metadata is **not** E2EE by
  definition (E2EE protects content, not who's-talking-to-whom — if you
  need to hide the social graph too, that's a much bigger scope/Tor-like
  problem; flag explicitly that this plan does not provide that).
- Ciphertext blobs (cannot decrypt without recipient's private key).

## MongoDB schema sketch

```
conversations
  _id
  participant_ids: [userId, craftsmanUserId]   // ref to Postgres users.id
  created_at
  last_message_at                              // for inbox sort

messages
  _id
  conversation_id
  sender_id
  created_at
  type: "text" | "image" | "video" | "audio"
  ciphertext: BinData                          // for text, inline
  nonce: BinData
  attachment:                                   // present for media types
    storage_url
    encrypted_key: BinData                      // envelope-encrypted per recipient
    content_type
    size_bytes
    checksum
  delivery_status: "sent" | "delivered" | "read"

device_keys
  user_id
  device_id
  identity_pubkey
  signed_prekey
  one_time_prekeys: [...]                       // consumed on use
```

Indexes: `conversations.participant_ids` (compound, for inbox lookups),
`messages.conversation_id + created_at` (for paginated thread fetch).

## Backend (Go) responsibilities

- New `internal/modules/messaging/` following the existing per-action
  module pattern (`send/`, `get_thread/`, `mark_read/`, `keys/`).
- New Mongo client wired into the existing `container.AppContainer` (mirror
  how `DB *gorm.DB` and `RDB *redis.Client` are already provided) — add
  `MongoDB *mongo.Database` alongside them; official `go.mongodb.org/mongo-driver`.
- Real-time delivery: add a WebSocket endpoint (Gin supports this via
  `gorilla/websocket` or `nhooyr.io/websocket`) for live message push;
  fall back to polling/REST `GET /messaging/conversations/:id/messages` for
  history. Redis (already provisioned) is a good fit for pub/sub fan-out if
  you need to support multiple API server instances later.
- JWT auth reused as-is for the messaging endpoints (existing middleware).
- Backend validates conversation membership (participant_ids contains
  caller's user_id) before allowing reads/writes, but never decrypts.

## Frontend (Angular) responsibilities

- New `services/messaging/` (key management, encrypt/decrypt, WebSocket
  client) and `components/messaging/` (conversation list, thread view,
  attachment capture/picker).
- WebCrypto API (`crypto.subtle`) or a vetted wrapper for all crypto
  operations — never implement AES/ECDH by hand.
- IndexedDB for private key storage (non-extractable keys where the API
  allows it).
- Capture flows for live photo/video/voice via `MediaDevices.getUserMedia`
  + `MediaRecorder`, encrypting the resulting Blob before upload.

## Phased rollout

1. **Phase 1 — text-only E2EE**: key exchange, Double Ratchet (or chosen
   library), text messages stored as ciphertext in Mongo, WebSocket
   delivery. Get the crypto primitives and key-storage UX right first.
2. **Phase 2 — attachments**: envelope encryption for files, reuse existing
   `Storage` abstraction, chunked encryption for larger media.
3. **Phase 3 — real-time capture**: in-app camera/mic capture UI on top of
   the Phase 2 attachment pipeline (no new crypto, just capture UX).
4. **Phase 4 — multi-device support**: if a user logs in from a second
   device, you need per-device key bundles and message fan-out to all of a
   recipient's devices (this is where Signal-protocol-style "sealed sender"
   /per-device sessions get genuinely complex — scope explicitly before
   committing to a timeline).

## Open decisions to make before starting implementation

- Build on an existing protocol/library (Signal protocol, Matrix
  Olm/Megolm) vs. rolling a simpler "session key exchanged via server,
  ratcheted by sequence number" scheme. Recommend the former for anything
  user-facing and security-sensitive; the latter only if this is a small
  side feature where reduced forward secrecy is acceptable.
- Where attachment blobs live in production (local filesystem, already
  used for products/profile pics, vs. object storage like S3/MinIO) — E2EE
  doesn't change this decision, but message volume/media size might push
  you toward object storage sooner than the rest of the app needed it.
- Whether group conversations are in scope (changes key exchange from
  pairwise X3DH to a group ratchet like Megolm/Signal Sender Keys) — confirm
  scope is 1:1 only before committing to X3DH+Double Ratchet, since adding
  groups later is a meaningful rework.
