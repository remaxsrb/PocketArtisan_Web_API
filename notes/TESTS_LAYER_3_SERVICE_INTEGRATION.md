# Layer 3 — Service Integration Tests

**Directory**: `tests/integration/order/`  
**Package**: `order_test`  
**Build tag**: `//go:build integration` on every file

These tests exercise `order/create`, `order/ship`, and `order/decline` services against a real PostgreSQL database. They are excluded from `go test ./...` by default and must be run explicitly.

## How to run

```
# Default (uses localhost postgres / pocketartisan_test db)
go test -tags integration ./tests/integration/order/... -v

# Custom DB
TEST_DATABASE_URL="host=my-host user=... dbname=..." \
  go test -tags integration ./tests/integration/order/... -v

# Custom assets dir (needed if run from outside repo root)
TEST_ASSETS_DIR=/path/to/assets \
  go test -tags integration ./tests/integration/order/... -v
```

## Prerequisites

1. A running PostgreSQL instance accessible via `TEST_DATABASE_URL` (defaults to `localhost:5432 / pocketartisan_test`).
2. Schema migrations already applied — the test DB must have the `orders`, `order_items`, `users`, and `products` tables.
3. Font assets at `TEST_ASSETS_DIR` (defaults to `../../../assets` relative to the test binary). Without fonts, `create` tests skip; `ship` and `decline` tests still run.

## Infrastructure (`setup_test.go`)

### `withTx`
Every test body runs inside a `testDB.Begin()` transaction registered for `t.Cleanup(tx.Rollback())`. All DB writes — seeded fixtures AND service-created rows — are invisible outside the transaction and are rolled back when the test finishes. The test DB stays empty between runs.

### `seedFixtures`
Creates a customer, a craftsman, and a product **inside the transaction**. No pre-existing data required.

### `freshMock`
Returns a new `payment.MockGateway` per test. Never share a mock between tests — `ForceNextError` is state that must not leak.

## Test inventory

### create_test.go (5 tests)

| Test | Scenario |
|------|----------|
| `TestCreate_COD_createsOrderPending` | COD order gets status `PENDING_CRAFTSMAN_REVIEW`, no reservation ID |
| `TestCreate_COD_totalPriceIsQuantityTimesUnitPrice` | Total = unit price × quantity for a COD order |
| `TestCreate_CC_reservesFundsAndPersistsReservationID` | CC order gets status `PAYMENT_RESERVED`; `PaymentReservationID` persisted on row |
| `TestCreate_CC_reserveFailure_compensatesWithOrderDeletion` | When Reserve fails, the already-committed order row is deleted (saga compensation) |
| `TestCreate_unknownProduct_returnsError` | Non-existent product ID returns an error; no order row created |

### ship_test.go (4 tests)

| Test | Scenario |
|------|----------|
| `TestShip_COD_doesNotCallCapture_statusShipped` | COD order ships without touching the payment gateway |
| `TestShip_CC_capturesReservation_statusShipped` | CC order: Capture called, reservation consumed, status → SHIPPED |
| `TestShip_CC_captureFailure_returnsError_statusUnchanged` | Capture fails → error returned, status stays unchanged |
| `TestShip_wrongCraftsman_returnsForbidden` | Craftsman ID mismatch → "forbidden" error |

### decline_test.go (4 tests)

| Test | Scenario |
|------|----------|
| `TestDecline_COD_doesNotCallRefund_statusDeclined` | COD order declined without touching the gateway |
| `TestDecline_CC_refundsReservation_statusDeclined` | CC order: Refund called, reservation consumed, status → DECLINED |
| `TestDecline_CC_refundFailure_returnsError_statusUnchanged` | Refund fails → error returned, status stays unchanged |
| `TestDecline_wrongCraftsman_returnsForbidden` | Craftsman ID mismatch → "forbidden" error |

## Mock data

Ship and decline tests insert orders directly via `tx.Create(&entities.Order{...})` rather than going through the create service. This keeps each test focused on the service under test and avoids any dependency on PDF generation or font assets.

For CC tests, a reservation is first created in `freshMock()` via `mock.Reserve(...)`. The returned ID is stored on the seeded order, mimicking what the create service would have done.

## Side effects

- `create` tests generate a PDF file in `t.TempDir()`. The temp directory is cleaned up automatically by the test runner — no manual cleanup needed.
- `BumpCacheVersion` is called by ship/decline services with `cache = nil`, which is a no-op (the function checks for nil cache).
