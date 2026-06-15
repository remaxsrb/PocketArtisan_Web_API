package create

import "PocketArtisan/internal/entities"

type NewOrderRequest struct {
	CraftsmanID     uint                  `json:"craftsman_id" binding:"required"`
	Items           []NewOrderItemRequest `json:"items" binding:"required,min=1,dive,required"`
	PaymentType     entities.PaymentType  `json:"payment_type" binding:"required"`
	ShippingAddress string                `json:"shipping_address" binding:"required"`
	CreditCardData  *CreditCardData       `json:"credit_card_data,omitempty"`
}

//additional data for customer will be fetched from the database using the customer ID, so we don't need to include it in the request. The same goes for craftsman data.

type NewOrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type CreditCardData struct {
	OwnerName      string `json:"owner_name" binding:"required"`
	CardNumber     string `json:"card_number" binding:"required"`
	ExpirationDate string `json:"expiration_date" binding:"required"`
	CVV            string `json:"cvv" binding:"required"`
}
