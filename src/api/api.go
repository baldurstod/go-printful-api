package api

import (
	"errors"
	"go-printful-api/src/model"
	"go-printful-api/src/model/requests"
	"go-printful-api/src/printful"
	"log"
	_ "net/http"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type ApiRequest struct {
	Action  string                 `json:"action" binding:"required"`
	Version int                    `json:"version" binding:"required"`
	Params  map[string]interface{} `json:"params"`
}

func ApiHandler(c *gin.Context) {
	var request ApiRequest
	var err error

	if err = c.ShouldBindJSON(&request); err != nil {
		log.Println(err)
		jsonError(c, errors.New("bad request"))
		return
	}

	switch request.Action {
	case "get-countries":
		err = getCountries(c)
	case "get-products":
		err = getProducts(c)
	case "get-product":
		err = getProduct(c, request.Params)
	case "get-product-prices":
		err = getProductPrices(c, request.Params)
	case "get-variant":
		err = getVariant(c, request.Params)
	case "get-similar-variants":
		err = getSimilarVariants(c, request.Params)
	case "get-templates":
		err = getTemplates(c, request.Params)
	case "get-styles":
		err = getStyles(c, request.Params)
	case "create-sync-product":
		err = createSyncProduct(c, request.Params)
	case "get-sync-product":
		err = getSyncProduct(c, request.Params)
	case "calculate-shipping-rates":
		err = calculateShippingRates(c, request.Params)
	case "calculate-tax-rate":
		err = calculateTaxRate(c, request.Params)
	case "create-order":
		err = createOrder(c, request.Params)
	default:
		jsonError(c, NotFoundError{})
		return
	}

	if err != nil {
		jsonError(c, err)
	}
}

func getCountries(c *gin.Context) error {
	countries, err := printful.GetCountries()

	if err != nil {
		return err
	}

	jsonSuccess(c, countries)

	return nil
}

func getProducts(c *gin.Context) error {
	products, err := printful.GetProducts()

	if err != nil {
		return err
	}

	jsonSuccess(c, products)

	return nil
}

func getProduct(c *gin.Context, params map[string]interface{}) error {
	productID := int(params["product_id"].(float64))
	product, err := printful.GetProduct(productID)

	if err != nil {
		return err
	}

	variants, err := printful.GetVariants(productID)

	if err != nil {
		return err
	}

	jsonSuccess(c, map[string]interface{}{
		"product":  product,
		"variants": variants,
	})

	return nil
}

func getProductPrices(c *gin.Context, params map[string]interface{}) error {
	productID := int(params["product_id"].(float64))
	currency := params["currency"].(string)

	prices, err := printful.GetProductPrices(productID, currency)

	if err != nil {
		return err
	}

	jsonSuccess(c, prices)

	return nil
}

func getVariant(c *gin.Context, params map[string]interface{}) error {
	variant, err := printful.GetVariant(int(params["variant_id"].(float64)))
	log.Println(params)

	if err != nil {
		return err
	}

	//log.Println("variant", variant)
	jsonSuccess(c, variant)

	return nil
}

func getSimilarVariants(c *gin.Context, params map[string]interface{}) error {

	log.Println("getSimilarVariants", params)

	p, placementOk := params["placement"]
	var placement string
	if !placementOk {
		placement = "default"
	} else {
		placement = p.(string)
	}

	variantIds, err := printful.GetSimilarVariants(int(params["variant_id"].(float64)), placement)
	log.Println(variantIds, err)

	jsonSuccess(c, variantIds)

	return nil
}

func getTemplates(c *gin.Context, params map[string]interface{}) error {
	templates, err := printful.GetMockupTemplates(int(params["product_id"].(float64)))
	log.Println(params)

	if err != nil {
		return err
	}

	jsonSuccess(c, templates)

	return nil
}

func getStyles(c *gin.Context, params map[string]interface{}) error {
	styles, err := printful.GetMockupStyles(int(params["product_id"].(float64)))
	log.Println(params)

	if err != nil {
		return err
	}

	jsonSuccess(c, styles)

	return nil
}

func createSyncProduct(c *gin.Context, params map[string]interface{}) error {
	createSyncProductRequest := model.CreateSyncProductDatas{}
	err := mapstructure.Decode(params, &createSyncProductRequest)
	if err != nil {
		log.Println(err)
		return errors.New("Error while decoding params")
	}

	syncProduct, err := printful.CreateSyncProduct(createSyncProductRequest)
	log.Println(syncProduct, err)

	jsonSuccess(c, syncProduct)

	return nil
}

func getSyncProduct(c *gin.Context, params map[string]interface{}) error {
	product, err := printful.GetSyncProduct(int64(params["sync_product_id"].(float64)))
	log.Println(product, params)

	if err != nil {
		return err
	}

	jsonSuccess(c, product)

	return nil
}

func calculateShippingRates(c *gin.Context, params map[string]interface{}) error {
	calculateShippingRatesRequest := requests.CalculateShippingRates{}
	err := mapstructure.Decode(params, &calculateShippingRatesRequest)
	if err != nil {
		log.Println(err)
		return errors.New("Error while decoding params")
	}

	shippingRates, err := printful.CalculateShippingRates(calculateShippingRatesRequest)
	log.Println(shippingRates, err)
	if err != nil {
		log.Println(err)
		return errors.New("Error while calculating shipping rates")
	}

	jsonSuccess(c, shippingRates)

	return nil
}

func calculateTaxRate(c *gin.Context, params map[string]interface{}) error {
	calculateTaxRateRequest := requests.CalculateTaxRate{}
	err := mapstructure.Decode(params, &calculateTaxRateRequest)
	if err != nil {
		log.Println(err)
		return errors.New("Error while decoding params")
	}

	shippingRates, err := printful.CalculateTaxRate(calculateTaxRateRequest)
	log.Println(shippingRates, err)
	if err != nil {
		log.Println(err)
		return errors.New("Error while calculating shipping rates")
	}

	jsonSuccess(c, shippingRates)

	return nil
}

func createOrder(c *gin.Context, params map[string]interface{}) error {
	log.Println("<<<<<<<<<<<<<<<<<<<<<<", params)

	createOrderRequest := requests.CreateOrderRequest{}
	err := mapstructure.Decode(params, &createOrderRequest)
	if err != nil {
		log.Println(err)
		return errors.New("Error while decoding params")
	}

	log.Println("=====================", createOrderRequest)

	order, err := printful.CreateOrder(createOrderRequest)
	log.Println(order, err)

	jsonSuccess(c, order)

	return nil
}
