package requests

import (
	printfulAPIModel "github.com/baldurstod/printful-api-model"
	"github.com/baldurstod/printful-api-model/schemas"
)

type CalculateShippingRates struct {
	Recipient printfulAPIModel.AddressInfo `mapstructure:"recipient"`
	Items     []printfulAPIModel.ItemInfo  `mapstructure:"items"`
	Currency  string                       `mapstructure:"currency"`
	Locale    string                       `mapstructure:"locale"`
}

type CalculateTaxRate struct {
	Recipient schemas.TaxAddressInfo `mapstructure:"recipient"`
}

type AddImagesRequest struct {
	Images []string `mapstructure:"images"`
}
