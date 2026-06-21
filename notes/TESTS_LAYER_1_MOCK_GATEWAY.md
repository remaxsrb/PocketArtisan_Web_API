# Layer 1 — MockGateway Unit Tests

**File**: `internal/modules/payment/mock_gateway_test.go`  
**Package**: `payment_test`  
**Build tag**: none (runs with `go test ./...`)

## What is tested

The `MockGateway` is a pure in-memory implementation of `payment.Gateway` used as a test double in all other layers. Layer 1 verifies the mock itself behaves correctly, so tests in Layers 2 and 3 can trust it.

| Test | Scenario |
|------|----------|
| `TestMockGateway_Reserve_returnsNonEmptyIDAndCorrectAmount` | Reserve returns a non-empty ID and echoes the requested amount |
| `TestMockGateway_Reserve_forceError_returnsErrAndClearsField` | `ForceNextError` injects a failure on the next call; field is cleared afterwards |
| `TestMockGateway_Reserve_multipleCallsProduceUniqueIDs` | Five sequential reservations each receive a distinct ID |
| `TestMockGateway_Capture_success_removesReservation` | Capture succeeds and removes the reservation; a second Capture on the same ID fails |
| `TestMockGateway_Capture_unknownID_returnsError` | Capture with an unknown ID returns an error |
| `TestMockGateway_Refund_success_removesReservation` | Refund succeeds and removes the reservation; a second Refund on the same ID fails |
| `TestMockGateway_Refund_unknownID_returnsError` | Refund with an unknown ID returns an error |
| `TestMockGateway_Reserve_concurrent_producesUniqueIDs` | 50 goroutines reserving concurrently all get unique IDs (data-race test) |

## How to run

```
go test ./internal/modules/payment/... -run TestMockGateway -v
```

Add `-race` to detect data races:

```
go test -race ./internal/modules/payment/... -run TestMockGateway -v
```

## Mock data

- Shared request fixture `testReserveReq` (OrderID=42, Amount=3200, Currency="RSD")
- No database, no filesystem, no network — everything is in memory

## Key design decisions

- `ForceNextError` is public and cleared after one use, so tests can inject exactly one failure without complex setup.
- Reservations are stored in a `map[string]Reservation` protected by a `sync.Mutex`, making the mock safe for the concurrency test.
- IDs are generated with `uuid.NewString()` (or equivalent) so uniqueness is guaranteed without a counter.
