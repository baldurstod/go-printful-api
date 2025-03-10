package requests

import (
	"github.com/baldurstod/printful-api-model/schemas"

	printfulsdk "github.com/baldurstod/go-printful-sdk/model"
)

type CalculateShippingRates struct {
	Recipient printfulsdk.ShippingRatesAddress                 `mapstructure:"recipient"`
	Items     []printfulsdk.CatalogOrWarehouseShippingRateItem `mapstructure:"items"`
	Currency  string                                           `mapstructure:"currency"`
	Locale    string                                           `mapstructure:"locale"`
}

type CalculateTaxRate struct {
	Recipient schemas.TaxAddressInfo `mapstructure:"recipient"`
}

type AddImagesRequest struct {
	Images []string `mapstructure:"images"`
}
