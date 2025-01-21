package model

type CreateSyncProductVariant struct {
	VariantID         int     `mapstructure:"variant_id"`
	ExternalVariantID string  `mapstructure:"external_variant_id"`
	RetailPrice       float64 `mapstructure:"retail_price"`
}
type CreateSyncProductDatas struct {
	ProductID int                        `mapstructure:"product_id"`
	Variants  []CreateSyncProductVariant `mapstructure:"variants"`
	Name      string                     `mapstructure:"name"`
	Image     string                     `mapstructure:"image"`
}
