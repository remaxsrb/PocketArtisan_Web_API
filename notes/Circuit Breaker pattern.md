
## Core Idea
Wraps unreliable operations (network calls, payment APIs, etc.) to monitor failures and **stop calling** a failing service instead of hammering it. It self-heals by periodically probing if the service has recovered.

---

## Three States

### CLOSED (Normal Operation)
- Requests flow through freely.
- Failure counter is tracked.
- If threshold exceeded → trips to **OPEN**.

### OPEN (Failing / Blocking)
- All requests are **immediately rejected** without calling the service.
- A timeout timer runs in the background.
- When timer expires → moves to **HALF-OPEN**.

### HALF-OPEN (Recovery Probe)
- Lets **one request** through as a probe.
- If probe succeeds → reset to **CLOSED**.
- If probe fails → revert to **OPEN**.

---

## Key Configuration
| Parameter | Description |
| :--- | :--- |
| `failureThreshold` | Number of failures required before opening the circuit. |
| `openTimeout` | Duration to stay open before attempting a probe. |
| `halfOpenProbes` | Number of concurrent probes allowed (typically 1). |

---

## State Machine Visualization

```text
                    failures >= threshold
      ┌─────────────────────────────────────────┐
      │                                         ▼
  [CLOSED]                                   [OPEN]
  requests flow                          all requests
  through freely         ◄───────────    fail fast
      ▲                                      │
      │                                      │ timeout
      │                                      │ expires
      │              probe succeeds          ▼
      └──────────────────────────────── [HALF-OPEN]
                                        one probe
                                        let through

```

## Operational Flows

### Failure Flow (CLOSED → OPEN)

req → breaker → service ✓   failure count: 0
req → breaker → service ✓   failure count: 0
req → breaker → service ✗   failure count: 1
req → breaker → service ✗   failure count: 2
req → breaker → service ✗   failure count: 3  ← threshold hit → OPEN
req → breaker ✗ (fast fail, service never called)

### Recovery Flow (OPEN → HALF-OPEN → CLOSED

[timeout expires]
req → breaker → service ✓   ← probe succeeds → CLOSED
req → breaker → service ✓   normal operation resumes

## What It Protects Against

Cascading failures: 

- Prevents errors from propagating across services.
- Resource exhaustion: Stops thread or goroutine leaks from hanging calls.
- Downstream overload: Prevents overwhelming a struggling service.
- Silent degradation: Forces immediate feedback rather than slow timeouts.