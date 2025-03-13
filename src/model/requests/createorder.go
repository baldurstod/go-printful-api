package requests

import (
	"github.com/baldurstod/go-printful-sdk/model"
)

type CreateOrderRequest struct {
	ExternalID    string               `json:"external_id" bson:"external_id" mapstructure:"external_id"`
	Shipping      string               `json:"shipping" bson:"shipping" mapstructure:"shipping"`
	Recipient     model.Address        `json:"recipient" bson:"recipient"`
	OrderItems    []model.CatalogItem  `json:"order_items" bson:"order_items"`
	Customization *model.Customization `json:"customization" bson:"customization" mapstructure:"customization"`
	RetailCosts   *model.RetailCosts2  `json:"retail_costs" bson:"retail_costs" mapstructure:"retail_costs"`
}
