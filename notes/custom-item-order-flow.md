# Custom Order Processing System Design

## Overview

This document describes the order processing workflow for a prototype marketplace platform that connects customers with craftsmen producing custom-made items.

The system supports two payment methods:

* **Cash on Delivery (COD)** – payment upon receiving the completed item.
* **Credit Card Payment** – using a **mock payment gateway** that simulates fund reservation and release. -> Open circut breaker pattern

Since this is a prototype, no real payment processor integration is implemented. The mock gateway is intended to demonstrate the business workflow and can later be replaced with a real payment provider.

---

# Order Lifecycle

## 1. Customer Places an Order

A customer browses the platform and selects one or more products offered by a craftsman.

During checkout, the customer submits an order containing:

* Ordered items
* Quantity of each item
* Delivery information
* Selected payment method

After successful submission, the order status becomes:

```
PENDING_CRAFTSMAN_REVIEW
```

The craftsman is notified that a new order requires attention.

---

## 2. Customer Selects Payment Method

The customer chooses one of the following payment options:

### Option A – Cash on Delivery

The customer agrees to pay after receiving the completed order.

Characteristics:

* No payment authorization occurs.
* Order proceeds directly to craftsman review.
* The craftsman produces and ships the item before receiving payment.

---

### Option B – Credit Card (Mock Gateway)

The customer chooses to pay using a credit card.

Since the system is a prototype, a mock payment gateway simulates the payment process.

Characteristics:

* The customer's funds are **reserved**.
* The craftsman does not immediately receive the money.
* The reserved funds are held until the order is shipped.

The order status becomes:

```
PAYMENT_RESERVED
```

---

# Credit Card Payment Flow

## 3. Fund Reservation

When the customer confirms payment:

1. The mock gateway reserves the required amount.
2. The reservation is recorded within the system.
3. The customer receives confirmation that the funds have been successfully reserved.

The money remains unavailable for transfer until the craftsman makes a decision regarding the order.

---

## 3.1 Craftsman Declines the Order

If the craftsman cannot fulfill the request:

* The craftsman provides a reason for declining the order.
* Optionally, the craftsman may suggest an alternative timeframe when they will be able to accept similar work.

Examples of decline reasons:

* Insufficient materials
* Excessive workload
* Custom request cannot be fulfilled
* Temporary unavailability

The customer receives:

* Notification that the order has been declined
* The explanation provided by the craftsman
* Any suggested future availability

### Refund Handling

If funds were previously reserved:

1. The mock gateway cancels the reservation.
2. The customer is notified that the reserved amount has been refunded.

The order status becomes:

```
DECLINED
```

---

# Craftsman Order Processing

## 4. Order PDF Generation

When a new order is received, the system generates a PDF document containing all information required for fulfillment.

The PDF may include:

* Order number
* Customer information
* Delivery address
* List of ordered items
* Quantities
* Product specifications
* Customization notes
* Selected payment method
* Order creation date

The generated PDF is delivered to the craftsman through the platform.

This document serves as the craftsman's work order.

---

# Shipment and Completion

## 5. Craftsman Ships the Order

After completing production, the craftsman marks the order as shipped.

The order status becomes:

```
SHIPPED
```

### For Credit Card Orders

The system instructs the mock payment gateway to capture the previously reserved funds.

The reserved amount is considered successfully collected.

### For Cash on Delivery Orders

No additional payment processing occurs within the system.

Payment is expected upon delivery.

---

# Receipt Generation

After shipment confirmation:

1. The system generates a receipt in PDF format.
2. The receipt is emailed to the customer.

The receipt may contain:

* Receipt number
* Order reference number
* Customer details
* Craftsman details
* Itemized list of products
* Quantities
* Total amount
* Payment method
* Shipment date

The receipt acts as proof of purchase for the customer.

---

# Order States

The following statuses describe the order lifecycle:

| Status                     | Description                                              |
| -------------------------- | -------------------------------------------------------- |
| `PENDING_CRAFTSMAN_REVIEW` | Order has been submitted and awaits craftsman action.    |
| `PAYMENT_RESERVED`         | Funds have been reserved through the mock gateway.       |
| `DECLINED`                 | Craftsman rejected the order.                            |
| `ACCEPTED`                 | Craftsman accepted the order and production has started. |
| `SHIPPED`                  | The completed order has been dispatched.                 |
| `COMPLETED`                | The order lifecycle has concluded successfully.          |

---

# Prototype Considerations

This implementation is intended solely for demonstration and validation purposes.

Current limitations include:

* No integration with real payment providers.
* Payment processing is simulated through a mock gateway.
* Shipping carrier integrations are not implemented.
* Email delivery may use a development SMTP service. https://github.com/rnwood/smtp4dev
* PDF generation focuses on functional requirements rather than legal compliance.

Future improvements may include:

* Integration with payment providers such as Stripe or PayPal.
* Automated shipment tracking.
* Customer order cancellation policies.
* Partial refunds and dispute handling.
* Digital signatures on generated receipts and invoices.

---

# Summary

The proposed design allows customers to order custom-made products while supporting both deferred and upfront payment scenarios.

The workflow emphasizes:

* Protection of customers through temporary fund reservation,
* Clear communication when craftsmen cannot fulfill orders,
* Automated document generation for craftsmen and customers, and
* A straightforward architecture suitable for a prototype that can evolve into a production-ready system.
