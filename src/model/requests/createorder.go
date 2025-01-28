package requests

import (
	"github.com/baldurstod/go-printful-sdk/model"
)

type CreateOrderRequest struct {
	Recipient model.Address       `json:"recipient" bson:"recipient"`
	Items     []model.CatalogItem `json:"items" bson:"items"`
}
