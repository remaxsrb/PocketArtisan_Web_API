package entities

import "time"

type OrderStatus string
type PaymentType string

const (
	OrderPending         OrderStatus = "PENDING_CRAFTSMAN_REVIEW"
	OrderPaymentReserved OrderStatus = "PAYMENT_RESERVED"
	OrderAccepted        OrderStatus = "ACCEPTED"
	OrderDeclined        OrderStatus = "DECLINED"
	OrderShipped         OrderStatus = "SHIPPED"
	OrderCompleted       OrderStatus = "COMPLETED"
)

const (
	PaymentCreditCard PaymentType = "CC"
	CashOnDelivery    PaymentType = "COD"
)

type Order struct {
	ID                   uint64      `json:"id" gorm:"primaryKey;autoIncrement"`
	CustomerID           uint64      `json:"customer_id"`
	CustomerAddress      string      `json:"customer_address"`
	CraftsmanID          uint64      `json:"craftsman_id"`
	TotalPrice           float64     `json:"total_price"`
	Items                []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	CreatedAt            time.Time   `json:"created_at" gorm:"autoCreateTime"`
	CompletedAt          *time.Time  `json:"completed_at,omitempty"`
	Status               OrderStatus `json:"status" gorm:"type:text;default:'PENDING_CRAFTSMAN_REVIEW'"`
	PaymentType          PaymentType `json:"payment_type" gorm:"type:text;'"`
	URL                  string      `json:"url" gorm:"type:text;" default:""`
	PaymentReservationID string      `json:"payment_reservation_id" gorm:"type:text;default:''"`
}

type OrderItem struct {
	ID        uint64  `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderID   uint64  `json:"order_id"`
	ProductID uint64  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price" gorm:"not null"`

	Order   Order   `json:"-" gorm:"foreignKey:OrderID"`
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}
