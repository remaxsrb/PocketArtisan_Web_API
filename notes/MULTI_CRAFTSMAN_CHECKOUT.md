# Multi-Craftsman Cart Checkout — Implementation Strategy

**Date**: June 17, 2026
**Status**: Proposed
**Scope**: New `cart/checkout` use case, reuse of `order/create` service, fund distribution per craftsman, impact on `ship` and `decline`

---

## Problem

The existing `POST /orders/create` accepts a single `CraftsmanID` and a flat list of items for that craftsman. A cart can hold products from N craftsmen. The current design cannot express this in one call.

---

## Core Design Decisions

### Checkout lives in the cart module

Checkout is the terminal action of the cart lifecycle: read cart → split by craftsman → create orders → clear cart. It is conceptually a cart operation, not an order operation. It belongs in `internal/modules/cart/checkout/`.

### Checkout delegates order creation to `order/create.Service`

Rather than duplicating the order-building logic, `cart/checkout` calls `order/create.Service.Execute` once per craftsman group. This keeps order creation in one place and checkout thin — it only orchestrates grouping, payment reservations, and cart clearing.

### Cart clearing moves out of `order/create`

Currently `order/create.Service.Execute` clears the cart inside its transaction. That was wrong placement — clearing the cart is a checkout concern, not an order concern. Remove it from `create` and move it to `cart/checkout`. `order/create` becomes a pure "create one order from these inputs" service with no cart awareness.

### One order per craftsman, one reservation per order (CC)

Each craftsman's slice of the cart becomes its own `Order` with its own `PaymentReservationID`. This keeps `ship` and `decline` self-contained — they only need to know about their own order's reservation, with no knowledge of the checkout that spawned it.

---

## Request Shape

`POST /cart/checkout` — client sends only payment metadata. Items come from the server-side cart.

```json
{
  "payment_type": "CC",
  "shipping_address": "Ulica Svetog Save 12, Beograd",
  "credit_card_data": {
    "owner_name": "Marko Marković",
    "card_number": "4111111111111111",
    "expiration_date": "12/27",
    "cvv": "123"
  }
}
```

`credit_card_data` is `omitempty` — omitted for COD.

```go
// internal/modules/cart/checkout/dto.go
package checkout

import "PocketArtisan/internal/entities"

type CheckoutRequest struct {
    PaymentType     entities.PaymentType `json:"payment_type" binding:"required"`
    ShippingAddress string               `json:"shipping_address" binding:"required"`
    CreditCardData  *CreditCardData      `json:"credit_card_data,omitempty"`
}

type CreditCardData struct {
    OwnerName      string `json:"owner_name" binding:"required"`
    CardNumber     string `json:"card_number" binding:"required"`
    ExpirationDate string `json:"expiration_date" binding:"required"`
    CVV            string `json:"cvv" binding:"required"`
}

type OrderResult struct {
    OrderID     uint64  `json:"order_id"`
    CraftsmanID uint64  `json:"craftsman_id"`
    Total       float64 `json:"total"`
    PDFURL      string  `json:"pdf_url,omitempty"`
}
```

---

## Service Logic (`cart/checkout/service.go`)

### Step 1 — Load the cart

Read `CartItem` rows for the authenticated user with `Product` preloaded (need `CraftsmanID` and `Price`). Fail early if the cart is empty or any product is unavailable.

### Step 2 — Group items by craftsman

```
[]CartItem → map[craftsmanID][]CartItem
```

### Step 3 — Call `order/create.Service.Execute` per group

For each craftsman group, build a `create.NewOrderRequest` from the group's items and call `orderCreateService.Execute(ctx, req)`. The `Execute` method:
- Snapshots prices from DB.
- Creates the `Order` and `OrderItem` rows in a transaction.
- Generates the PDF.
- Returns the PDF URL.

It does **not** clear the cart (that responsibility is removed — see below).

```go
for craftsmanID, items := range groups {
    req := create.NewOrderRequest{
        CraftsmanID:     craftsmanID,
        Items:           toOrderItems(items),
        PaymentType:     checkoutReq.PaymentType,
        ShippingAddress: checkoutReq.ShippingAddress,
    }
    result, err := uc.orderCreate.Execute(ctx, req)
    if err != nil {
        // compensate previously reserved orders and return
        rollbackReservations(ctx, uc.gateway, reservedSoFar)
        deleteCreatedOrders(ctx, uc.db, createdOrderIDs)
        return nil, err
    }
    // ... handle CC reservation (Step 4)
}
```

`orderCreate` is stored on the `checkout.Service` struct as `*create.Service`.

### Step 4 — Reserve funds per order (CC only)

After each `Execute` succeeds, call `gateway.Reserve` for that order. If any reservation fails, refund all prior ones and delete the DB orders created so far (manual saga):

```go
if req.PaymentType == entities.PaymentCreditCard {
    res, err := uc.gateway.Reserve(ctx, payment.ReserveRequest{
        OrderID: createdOrderID,
        Amount:  orderTotal,
        Currency: "RSD",
    })
    if err != nil {
        rollbackReservations(ctx, uc.gateway, reservedSoFar)
        deleteCreatedOrders(ctx, uc.db, createdOrderIDs)
        return nil, fmt.Errorf("payment reservation failed: %w", err)
    }
    reservedSoFar = append(reservedSoFar, res.ID)
    uc.db.Model(&entities.Order{}).Where("id = ?", createdOrderID).
        Update("payment_reservation_id", res.ID)
}
```

### Step 5 — Clear the cart

Only after all orders are created and all reservations succeed:

```go
uc.db.Where("cart_id = (SELECT id FROM carts WHERE user_id = ?)", customerID).
    Delete(&entities.CartItem{})
uc.db.Model(&entities.Cart{}).Where("user_id = ?", customerID).
    Update("total", 0)
```

### Step 6 — Return `[]OrderResult`

One entry per craftsman order with `OrderID`, `CraftsmanID`, `Total`, and `PDFURL`.

---

## Changes to `order/create/service.go`

Remove the cart-clearing block from `Execute`. Everything else stays identical. The method signature and behaviour from the caller's perspective is unchanged — it still creates one order and returns a PDF URL.

```go
// DELETE from order/create/service.go — move to cart/checkout:
if err := tx.Where("cart_id = (SELECT id FROM carts WHERE user_id = ?)", customer.ID).
    Delete(&entities.CartItem{}).Error; err != nil { ... }
if err := tx.Model(&entities.Cart{}).Where("user_id = ?", customer.ID).
    Update("total", 0).Error; err != nil { ... }
```

`POST /orders/create` continues to work exactly as before for single-craftsman use, now without the side effect of clearing the cart.

---

## Impact on `ship` and `decline`

### `order/ship/service.go`

Inject `payment.Gateway`. On ship, call `Capture` if the order is CC:

```go
if existing.PaymentType == entities.PaymentCreditCard && existing.PaymentReservationID != "" {
    if err := uc.gateway.Capture(ctx, existing.PaymentReservationID); err != nil {
        return "", fmt.Errorf("capture payment: %w", err)
    }
}
```

### `order/decline/service.go`

Inject `payment.Gateway`. On decline, call `Refund` if the order is CC:

```go
if existing.PaymentType == entities.PaymentCreditCard && existing.PaymentReservationID != "" {
    if err := uc.gateway.Refund(ctx, existing.PaymentReservationID); err != nil {
        return "", fmt.Errorf("refund payment: %w", err)
    }
}
```

---

## Entity Change

Add `PaymentReservationID` to `Order` (also required by `PAYMENT_GATEWAY_MOCK_AND_TESTING_PLAN.md`):

```go
// internal/entities/order.go
type Order struct {
    // ... existing fields
    PaymentReservationID string `json:"payment_reservation_id" gorm:"type:text;default:''"`
}
```

Empty string for COD orders and all existing rows.

---

## Routing

```go
// internal/http/routes/cart_routes.go
checkout.RegisterRoutes(customerCartRoutes, appContainer.DB, appContainer.RDB, appContainer.Storage, appContainer.Fonts, appContainer.Gateway)

// internal/http/routes/order_routes.go — only these two lines change:
ship.RegisterRoutes(craftsmanOrderRoutes, appContainer.DB, appContainer.RDB, appContainer.Gateway)
decline.RegisterRoutes(craftsmanOrderRoutes, appContainer.DB, appContainer.RDB, appContainer.Gateway)
```

---

## File Map

```
internal/modules/cart/
    checkout/
        dto.go        ← CheckoutRequest, CreditCardData, OrderResult
        service.go    ← cart read, grouping, calls order/create per group, reservations, cart clear
        controller.go ← POST /cart/checkout

internal/modules/order/
    create/
        service.go    ← remove cart-clearing block (only change)
    ship/
        service.go    ← add Capture call for CC orders
    decline/
        service.go    ← add Refund call for CC orders
```

---

## Scenarios to Validate

| # | Scenario | Expected |
|---|----------|----------|
| 1 | COD checkout, 2 craftsmen | `order/create.Execute` called twice; cart cleared; `Reserve` never called |
| 2 | CC checkout, 2 craftsmen | `order/create.Execute` called twice; `Reserve` called twice; each order has its own `PaymentReservationID`; cart cleared |
| 3 | CC checkout, second `Reserve` fails | First order deleted; first reservation refunded; cart unchanged |
| 4 | `order/create.Execute` fails for one craftsman | No orders committed; cart unchanged |
| 5 | Cart is empty at checkout | 400 before any order or reservation logic runs |
| 6 | `POST /orders/create` called directly | Works as before; cart is NOT cleared (side effect removed) |
| 7 | Craftsman ships CC order | `Capture` called with that order's `PaymentReservationID` only |
| 8 | Craftsman declines CC order | `Refund` called with that order's `PaymentReservationID` only |
| 9 | Craftsman ships COD order | `Capture` never called |

---

## What Stays Unchanged

- `order/accept` — no payment interaction, untouched.
- `order/get_by_customer`, `order/get_by_craftsman` — untouched.
- `cart/add_to_cart`, `cart/remove_from_cart` — untouched.
- PDF generation internals — reused as-is via `order/create`.
- All existing `order/create` callers and tests — behaviour unchanged except cart is no longer cleared as a side effect.
