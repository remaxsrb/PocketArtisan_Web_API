# Layer 2 — BreakerGateway Unit Tests

**File**: `internal/modules/payment/breaker_gateway_test.go`  
**Package**: `payment_test`  
**Build tag**: none (runs with `go test ./...`)

## What is tested

`BreakerGateway` wraps any `payment.Gateway` and implements the circuit-breaker pattern with three states: CLOSED → OPEN → HALF_OPEN → CLOSED. Layer 2 tests all state transitions and the concurrency invariant (only one probe reaches the inner gateway while HALF_OPEN).

| Test | Scenario |
|------|----------|
| `TestBreakerGateway_CLOSED_requestsFlowThrough` | Five successive calls all reach the inner gateway |
| `TestBreakerGateway_CLOSED_failuresBelowThreshold_staysClosed` | Two failures out of three threshold do not trip the breaker |
| `TestBreakerGateway_CLOSED_tripsToOPEN_atThreshold` | Three failures trip the breaker; next call returns "circuit open" without calling inner |
| `TestBreakerGateway_OPEN_fastFailWithoutCallingInner` | Five calls while OPEN all fast-fail; inner call count is unchanged |
| `TestBreakerGateway_OPEN_transitionsToHALFOPEN_afterTimeout` | After the open timeout elapses the first call is allowed through as a probe |
| `TestBreakerGateway_HALFOPEN_probeSuccess_resetsToCLOSED` | A successful probe resets the breaker to CLOSED; subsequent calls flow freely |
| `TestBreakerGateway_HALFOPEN_probeFailure_reopensImmediately` | A failed probe immediately re-opens the breaker (no re-tripping threshold) |
| `TestBreakerGateway_HALFOPEN_concurrent_onlyOneProbeReachesInner` | While a probe is in-flight, concurrent callers get "probe in flight" error |
| `TestBreakerGateway_CLOSED_successResetsFailureCounter` | A success resets the failure counter so the threshold window starts fresh |

## How to run

```
go test ./internal/modules/payment/... -run TestBreakerGateway -v
```

With race detector (recommended — tests use goroutines):

```
go test -race ./internal/modules/payment/... -run TestBreakerGateway -v
```

## Test helpers

### `countingGateway`
Wraps `MockGateway` and keeps an atomic call counter on `Reserve`. Used to assert that OPEN/HALF_OPEN states do not forward calls to the inner gateway.

### `stateGateway`
Supports three runtime modes set via `setMode`:
- `"fail"` — returns an error immediately (used to trip the breaker)
- `"block"` — signals on `entered` channel, then waits for `release` to be closed (used to hold a probe open so a second goroutine can race it)
- anything else — returns success

### `tripBreaker(b, cg, n)`
Injects `ForceNextError` and calls Reserve `n` times to trip the breaker to OPEN in a single helper call.

### `newBreaker(inner)`
Creates a breaker with `threshold=3` and `timeout=30ms` — short enough that recovery tests are not flaky.

## Key design decisions

The HALF_OPEN concurrency test is the most subtle: the `stateGateway` in "block" mode lets the test pause the probe goroutine inside `inner.Reserve`, then assert that a second goroutine gets a "probe in flight" error before releasing the probe. This proves `probeInFlight bool` guards correctly under concurrent access.

The 30 ms open timeout and `time.Sleep(40ms)` in recovery tests add 10 ms margin for scheduling variance. This is intentional — tighter margins cause flaky tests on loaded CI runners.
