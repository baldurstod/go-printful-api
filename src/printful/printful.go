package printful

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	printfulsdk "github.com/baldurstod/go-printful-sdk"
	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
	printfulAPIModel "github.com/baldurstod/printful-api-model"

	"github.com/baldurstod/go-printful-api-model/requests"
	"github.com/baldurstod/printful-api-model/responses"
	"github.com/baldurstod/printful-api-model/schemas"

	"bytes"
	"encoding/base64"
	"go-printful-api/src/config"
	"go-printful-api/src/model"
	"go-printful-api/src/mongo"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baldurstod/randstr"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/image/draw"
)

var printfulConfig config.Printful
var printfulClient *printfulsdk.PrintfulClient = printfulsdk.NewPrintfulClient("")

func SetPrintfulConfig(config config.Printful) {
	printfulConfig = config
	log.Println(config)
	printfulClient.SetAccessToken(config.AccessToken)
	//go initAllProducts()
}

const PRINTFUL_PRODUCTS_API = "https://api.printful.com/products"
const PRINTFUL_STORE_API = "https://api.printful.com/store"
const PRINTFUL_MOCKUP_GENERATOR_API = "https://api.printful.com/mockup-generator"
const PRINTFUL_MOCKUP_GENERATOR_API_CREATE_TASK = "https://api.printful.com/mockup-generator/create-task"
const PRINTFUL_COUNTRIES_API = "https://api.printful.com/countries"
const PRINTFUL_ORDERS_API = "https://api.printful.com/orders"
const PRINTFUL_SHIPPING_API = "https://api.printful.com/shipping"
const PRINTFUL_TAX_API = "https://api.printful.com/tax"

var mutexPerEndpoint = make(map[string]*sync.Mutex)

func addEndPoint(endPoint string) int {
	mutexPerEndpoint[endPoint] = &sync.Mutex{}
	return 0
}

var _ = addEndPoint(PRINTFUL_PRODUCTS_API)
var _ = addEndPoint(PRINTFUL_STORE_API)
var _ = addEndPoint(PRINTFUL_MOCKUP_GENERATOR_API)
var _ = addEndPoint(PRINTFUL_MOCKUP_GENERATOR_API_CREATE_TASK)
var _ = addEndPoint(PRINTFUL_COUNTRIES_API)
var _ = addEndPoint(PRINTFUL_ORDERS_API)
var _ = addEndPoint(PRINTFUL_SHIPPING_API)
var _ = addEndPoint(PRINTFUL_TAX_API)

func fetchRateLimited(method string, apiURL string, path string, headers map[string]string, body map[string]interface{}) (*http.Response, error) {
	mutex := mutexPerEndpoint[apiURL]

	mutex.Lock()
	defer mutex.Unlock()

	u, err := url.JoinPath(apiURL, path)
	if err != nil {
		return nil, errors.New("unable to create URL")
	}

	var requestBody io.Reader
	if body != nil {
		out, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(out)
	}

	var resp *http.Response
	for i := 0; i < 10; i++ {

		req, err := http.NewRequest(method, u, requestBody)
		if err != nil {
			return nil, err
		}

		for k, v := range headers {
			req.Header.Add(k, v)
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 { //Too Many Requests
			time.Sleep(60 * time.Second)
			continue
		}

		if resp.StatusCode != 200 { //Everything except 429 and 200
			return nil, fmt.Errorf("printful returned HTTP status code: %d", resp.StatusCode)
		}
		break
	}

	header := resp.Header
	remaining := header.Get("X-RateLimit-Remaining")
	if remaining == "" {
		return resp, err
	}

	remain, err := strconv.Atoi(remaining)
	if err != nil {
		return nil, errors.New("unable to get rate limit")
	}

	reset, err := strconv.Atoi(header.Get("X-Ratelimit-Reset"))
	if err != nil {
		reset = 60 // default to 60s
		err = nil
	}

	if remain < 1 {
		reset += 2
		time.Sleep(time.Duration(reset) * time.Second)
	}

	return resp, err
}

func RefreshAllProducts(currency string, useCache bool) error {
	products, err := printfulClient.GetCatalogProducts()
	if err != nil {
		return errors.New("unable to get printful response")
	}

	for _, product := range products {
		mongo.InsertProduct(&product)
	}

	for _, product := range products {
		if err = refreshVariants(product.ID, product.VariantCount, useCache); err != nil {
			log.Println("Error while refreshing product variants", product.ID, err)
		}

		if err = refreshPrices(product.ID, currency, useCache); err != nil {
			log.Println("Error while refreshing product prices", product.ID, err)
		}

		if err = refreshTemplates(product.ID, useCache); err != nil {
			log.Println("Error while refreshing product templates", product.ID, err)
		}

		if err = refreshStyles(product.ID, useCache); err != nil {
			log.Println("Error while refreshing product styles", product.ID, err)
		}
	}

	return nil
}

func refreshVariants(productID int, count int, useCache bool) error {
	//log.Println("Refreshing variants for product", productID)

	var variants []printfulmodel.Variant
	outdated := true
	var err error

	if useCache {
		variants, outdated, err = mongo.FindVariants(productID)
		if err != nil || len(variants) != count {
			outdated = true
		}
	}

	if outdated {
		log.Println("Variants for product", productID, "are outdated, refreshing")
		variants, err = printfulClient.GetCatalogVariants(productID)
		if err != nil {
			//log.Println("Error while getting product variants", productID, err)
			return fmt.Errorf("error in refreshVariants: %w", err)
		} else {

			variantIDs := make([]int, 0, 20)

			for _, variant := range variants {
				variantIDs = append(variantIDs, variant.ID)
				if err = mongo.InsertVariant(&variant); err != nil {
					return fmt.Errorf("error in refreshVariants: %w", err)
				}
			}

			if err = mongo.UpdateProductVariantIds(productID, variantIDs); err != nil {
				return fmt.Errorf("error in refreshVariants: %w", err)

			}
		}
	}
	return nil
}

func refreshPrices(productID int, currency string, useCache bool) error {
	var prices *printfulmodel.ProductPrices
	outdated := true
	var err error

	if useCache {
		_, outdated, err = mongo.FindProductPrices(productID, currency)
		if err != nil {
			outdated = true
		}
	}

	if outdated {
		log.Println("Prices for product", productID, "currency", currency, "are outdated, refreshing")
		prices, err = printfulClient.GetProductPrices(productID)
		if err != nil {
			return fmt.Errorf("error in refreshPrices: %w", err)
		} else {
			mongo.InsertProductPrices(prices)
		}
	}

	return nil
}

func refreshTemplates(productID int, useCache bool) error {
	var templates []printfulmodel.MockupTemplates
	outdated := true
	var err error

	if useCache {
		_, outdated, err = mongo.FindMockupTemplates(productID)
		if err != nil {
			outdated = true
		}
	}

	if outdated {
		log.Println("Templates for product", productID, "are outdated, refreshing")
		templates, err = printfulClient.GetMockupTemplates(productID)
		if err != nil {
			return fmt.Errorf("error in refreshTemplates: %w", err)
		} else {
			mongo.InsertMockupTemplates(productID, templates)
		}
	}

	return nil
}

func refreshStyles(productID int, useCache bool) error {
	var styles []printfulmodel.MockupStyles
	outdated := true
	var err error

	if useCache {
		_, outdated, err = mongo.FindMockupStyles(productID)
		if err != nil {
			outdated = true
		}
	}

	if outdated {
		log.Println("Styles for product", productID, "are outdated, refreshing")
		styles, err = printfulClient.GetMockupStyles(productID)
		if err != nil {
			return fmt.Errorf("error in refreshStyles: %w", err)
		} else {
			mongo.InsertMockupStyles(productID, styles)
		}
	}

	return nil
}

func GetCategories() ([]printfulmodel.Category, error) {
	categories, err := mongo.FindCategories()

	if err != nil {
		return nil, err
	}

	return categories, nil
}

type GetCountriesResponse struct {
	Code   int                        `json:"code"`
	Result []printfulAPIModel.Country `json:"result"`
}

func GetCountries() ([]printfulmodel.Country, error) {
	countries, err := mongo.FindCountries()

	if err != nil {
		return nil, err
	}

	return countries, nil
}

type GetProductsResponse struct {
	Code   int                        `json:"code"`
	Result []printfulAPIModel.Product `json:"result"`
}

//var cachedProducts = make([]printfulmodel.Product, 0)
//var cachedProductsUpdated = time.Time{}

func GetProducts() ([]printfulmodel.Product, error) {
	products, err := mongo.FindProducts()

	if err != nil {
		return nil, err
	}

	return products, nil
}

type GetProductResponse struct {
	Code   int                          `json:"code"`
	Result printfulAPIModel.ProductInfo `json:"result"`
}

func GetProduct(productID int) (*printfulmodel.Product, error) {
	product, _, err := mongo.FindProduct(productID)
	if err == nil {
		return product, nil
	}

	return nil, errors.New("unable to find product")
}

func GetProductPrices(productID int, currency string) (*printfulmodel.ProductPrices, error) {
	productPrices, _, err := mongo.FindProductPrices(productID, currency)
	if err != nil {
		return nil, errors.New("unable to find product prices")
	}

	for i := range productPrices.Product.Placements {
		placement := &productPrices.Product.Placements[i]

		placement.Price, err = applyMarkup(placement.Price, printfulConfig.Markup) //(1 + printfulConfig.Markup/100)
		if err != nil {
			return nil, errors.New("failed to format product price")
		}

		placement.DiscountedPrice, err = applyMarkup(placement.DiscountedPrice, printfulConfig.Markup) //(1 + printfulConfig.Markup/100)
		if err != nil {
			return nil, errors.New("failed to format product price")
		}
	}

	for i := range productPrices.Variants {
		variant := &productPrices.Variants[i]
		for j := range variant.Techniques {
			technique := &variant.Techniques[j]
			technique.Price, err = applyMarkup(technique.Price, printfulConfig.Markup) //(1 + printfulConfig.Markup/100)
			if err != nil {
				return nil, errors.New("failed to format product price")
			}

			technique.DiscountedPrice, err = applyMarkup(technique.DiscountedPrice, printfulConfig.Markup) //(1 + printfulConfig.Markup/100)
			if err != nil {
				return nil, errors.New("failed to format product price")
			}
		}
	}

	return productPrices, nil
}

func applyMarkup(price string, pct float64) (string, error) {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return "", err
	}

	p *= (1 + pct*0.01)
	return strconv.FormatFloat(p, 'f', 2, 64), nil
}

func GetVariants(productID int) ([]printfulmodel.Variant, error) {
	variants, _, err := mongo.FindVariants(productID)
	if err == nil {
		return variants, nil
	}

	return nil, errors.New("unable to find variants")
}

type GetVariantResponse struct {
	Code   int                          `json:"code"`
	Result printfulAPIModel.VariantInfo `json:"result"`
}

func GetVariant(variantID int) (*printfulmodel.Variant, error) {
	variant, _, err := mongo.FindVariant(variantID)
	if err == nil {
		return variant, nil
	}

	return nil, errors.New("unable to find variant")
}

type GetTemplatesResponse struct {
	Code   int                              `json:"code"`
	Result printfulAPIModel.ProductTemplate `json:"result"`
}

func GetMockupTemplates(productID int) ([]printfulmodel.MockupTemplates, error) {
	templates, _, err := mongo.FindMockupTemplates(productID)

	if err != nil {
		return nil, err
	}

	return templates, nil
}

func GetMockupStyles(productID int) ([]printfulmodel.MockupStyles, error) {
	styles, _, err := mongo.FindMockupStyles(productID)

	if err != nil {
		return nil, err
	}

	return styles, nil
}

type GetSimilarVariantsPlacement struct {
	Placement   string `json:"placement"`
	Technique   string `json:"technique"`
	Orientation string `json:"orientation"`
}

func GetSimilarVariants(variantID int, placements []GetSimilarVariantsPlacement) ([]int, error) {
	if placements == nil {
		return nil, errors.New("placement is nil")
	}

	variant, err := GetVariant(variantID)
	if err != nil {
		return nil, err
	}

	product, err := GetProduct(variant.CatalogProductID)
	if err != nil {
		return nil, err
	}

	templates, err := GetMockupTemplates(variant.CatalogProductID)

	variantsIDs := make(map[int]int, 0)

	for _, v := range product.CatalogVariantIDs {
		if (variantID == v) || matchTemplate(templates, variantID, v, placements) {
			variantsIDs[v] = v
		}
	}

	keys := make([]int, len(variantsIDs))
	i := 0
	for k := range variantsIDs {
		keys[i] = k
		i++
	}

	return keys, nil
}

func matchTemplate(templates []printfulmodel.MockupTemplates, v1 int, v2 int, placements []GetSimilarVariantsPlacement) bool {
	for _, placement := range placements {
		template1 := findTemplate(templates, v1, &placement)
		if template1 == nil {
			return false
		}

		template2 := findTemplate(templates, v2, &placement)
		if template2 == nil {
			return false
		}

		if template1.PrintAreaWidth == 0 || template1.PrintAreaHeight == 0 {
			return false
		}

		if template1.TemplateWidth != template2.TemplateWidth ||
			template1.TemplateHeight != template2.TemplateHeight ||
			template1.PrintAreaWidth != template2.PrintAreaWidth ||
			template1.PrintAreaHeight != template2.PrintAreaHeight ||
			template1.PrintAreaTop != template2.PrintAreaTop ||
			template1.PrintAreaLeft != template2.PrintAreaLeft {
			return false
		}
	}
	return true
}

func findTemplate(templates []printfulmodel.MockupTemplates, variantID int, placement *GetSimilarVariantsPlacement) *printfulmodel.MockupTemplates {
	idx := slices.IndexFunc(templates, func(t printfulmodel.MockupTemplates) bool {
		if t.Orientation != placement.Orientation ||
			t.Technique != placement.Technique ||
			t.Placement != placement.Placement {
			return false
		}

		idx := slices.IndexFunc(t.CatalogVariantIDs, func(id int) bool { return id == variantID })
		if idx == -1 {
			return false
		}

		return true
	})

	if idx == -1 {
		return nil
	}

	return &templates[idx]
}

func matchPrintFile(printfileInfo *printfulAPIModel.PrintfileInfo, variantID1 int, variantID2 int, placement string) bool {
	//log.Println(printfileInfo)
	printfile1 := printfileInfo.GetPrintfile(variantID1, placement)
	printfile2 := printfileInfo.GetPrintfile(variantID2, placement)

	if (printfile1 != nil) && (printfile2 != nil) {
		return (printfile1.Width == printfile2.Width) && (printfile1.Height == printfile2.Height)
	}
	return false
}

type CreateSyncProductResponse struct {
	Code   int                 `json:"code"`
	Result schemas.SyncProduct `json:"result"`
}

func CreateSyncProduct(datas model.CreateSyncProductDatas) (*schemas.SyncProduct, error) {
	//log.Println("CreateSyncProduct", datas)

	b64data := datas.Image[strings.IndexByte(datas.Image, ',')+1:] // Remove data:image/png;base64,

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64data))

	config, err := png.DecodeConfig(reader)
	if err != nil {
		return nil, err
	}

	if config.Width > 20000 || config.Height > 20000 {
		return nil, errors.New("image too large")
	}

	img, err := png.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64data)))
	if err != nil {
		return nil, err
	}

	newWidth, newHeight := 200, 200
	scaledImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	srcRectangle := img.Bounds()
	dstRectangle := scaledImage.Bounds()

	scrWidth := srcRectangle.Dx()
	scrHeigh := srcRectangle.Dy()

	log.Println(scrWidth, scrHeigh)
	//dstWidth := dstRectangle.Dx()
	//dstHeigh := dstRectangle.Dy()

	srcRatio := float64(scrWidth) / float64(scrHeigh)

	if srcRatio > 1 {
		// width > heigh
		h := int(float64(newHeight) / srcRatio)
		dstRectangle.Min.Y = (newHeight - h) / 2
		dstRectangle.Max.Y = dstRectangle.Min.Y + h
	} else if srcRatio < 1 {
		// heigh > width
		w := int(float64(newWidth) * srcRatio)
		dstRectangle.Min.X = (newWidth - w) / 2
		dstRectangle.Max.X = dstRectangle.Min.X + w
		log.Println(dstRectangle)
	}

	draw.CatmullRom.Scale(scaledImage, dstRectangle, img, srcRectangle, draw.Over, nil)

	filename := randstr.String(32)
	log.Println(filename)

	err = mongo.UploadImage(filename, img)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = mongo.UploadImage(filename+"_thumb", scaledImage)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	headers := map[string]string{
		"Authorization": "Bearer " + printfulConfig.AccessToken,
	}

	imageURL, err := url.JoinPath(printfulConfig.ImagesURL, "/", filename)
	if err != nil {
		return nil, errors.New("unable to create image url")
	}

	syncVariants := []map[string]interface{}{}
	for _, v := range datas.Variants {
		syncVariant := map[string]interface{}{
			"variant_id":   v.VariantID,
			"external_id":  v.ExternalVariantID,
			"retail_price": v.RetailPrice,
			"files": []interface{}{
				map[string]interface{}{
					"url": imageURL,
				},
			},
		}
		syncVariants = append(syncVariants, syncVariant)
	}

	thumbnailURL, err := url.JoinPath(printfulConfig.ImagesURL, "/", filename+"_thumb")
	if err != nil {
		return nil, errors.New("unable to create thumbnail url")
	}

	body := map[string]interface{}{
		"sync_product": map[string]interface{}{
			"name":      datas.Name,
			"thumbnail": thumbnailURL,
		},
		"sync_variants": syncVariants,
	}

	log.Println(body)

	resp, err := fetchRateLimited("POST", PRINTFUL_STORE_API, "/products", headers, body)
	if err != nil {
		return nil, errors.New("unable to get printful response")
	}

	response := CreateSyncProductResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to decode printful response")
	}

	log.Println(response)

	p := &(response.Result)

	return p, nil
}

type GetSyncProductResponse struct {
	Code   int                              `json:"code"`
	Result printfulAPIModel.SyncProductInfo `json:"result"`
}

func GetSyncProduct(syncProductID int64) (*printfulAPIModel.SyncProductInfo, error) {
	/*product, err := mongo.FindProduct(productID)
	if err == nil {
		return product, nil, false
	}*/
	headers := map[string]string{
		"Authorization": "Bearer " + printfulConfig.AccessToken,
	}

	resp, err := fetchRateLimited("GET", PRINTFUL_STORE_API, "/products/"+strconv.FormatInt(syncProductID, 10), headers, nil)
	if err != nil {
		return nil, errors.New("unable to get printful response")
	}

	//body, _ := ioutil.ReadAll(resp.Body)
	//log.Println(string(body))
	response := GetSyncProductResponse{}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to decode printful response")
	}

	if response.Code != 200 {
		log.Println(err)
		return nil, errors.New("printful returned an error")
	}

	p := &(response.Result)
	log.Println(response.Result)

	return p, nil
}

func CalculateShippingRates(datas requests.CalculateShippingRates) ([]printfulmodel.ShippingRate, error) {
	shippingRates, err := printfulClient.CalculateShippingRates(datas.Recipient, datas.Items, printfulsdk.WithCurrency(datas.Currency), printfulsdk.WithLanguage(datas.Locale))
	if err != nil {
		return nil, errors.New("unable to get printful response")
	}

	return shippingRates, nil
}

func CalculateTaxRate(datas requests.CalculateTaxRate) (*schemas.TaxInfo, error) {
	body := map[string]interface{}{}
	err := mapstructure.Decode(datas, &body)
	if err != nil {
		log.Println(err)
		return nil, errors.New("error while decoding params")
	}

	log.Println(body)

	headers := map[string]string{
		"Authorization": "Bearer " + printfulConfig.AccessToken,
	}

	resp, err := fetchRateLimited("POST", PRINTFUL_TAX_API, "/rates", headers, body)
	if err != nil {
		return nil, errors.New("unable to get printful response")
	}
	defer resp.Body.Close()

	//response := map[string]interface{}{}
	response := responses.TaxRates{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to decode printful response")
	}
	log.Println(response)

	//p := &(response.Result)

	return &response.Result, nil
}

type CreateOrderResponse struct {
	Code   int           `json:"code"`
	Result schemas.Order `json:"result"`
}

func CreateOrder(request requests.CreateOrder) (*printfulmodel.Order, error) {
	opts := make([]printfulsdk.RequestOption, 0, 5)

	if request.ExternalID != "" {
		opts = append(opts, printfulsdk.SetOrderExternalID(request.ExternalID))
	}

	if request.Shipping != "" {
		opts = append(opts, printfulsdk.SetOrderShippingMethod(request.Shipping))
	}

	if request.Customization != nil {
		opts = append(opts, printfulsdk.SetOrderCustomization(request.Customization))
	}

	if request.RetailCosts != nil {
		opts = append(opts, printfulsdk.SetOrderRetailCosts(request.RetailCosts))
	}

	/*

		ExternalID    string              `json:"external_id" bson:"external_id" mapstructure:"external_id"`
		Shipping      string              `json:"shipping" bson:"shipping" mapstructure:"shipping"`
		Recipient     model.Address       `json:"recipient" bson:"recipient"`
		OrderItems    []model.CatalogItem `json:"order_items" bson:"order_items"`
		Customization model.Customization `json:"customization" bson:"customization" mapstructure:"customization"`
		RetailCosts   model.RetailCosts2  `json:"retail_costs" bson:"retail_costs" mapstructure:"retail_costs"`
	*/

	order, err := printfulClient.CreateOrder(request.Recipient, request.OrderItems, opts...)
	if err != nil {
		return nil, errors.New("unable to get printful response")
	}

	return order, nil

	/*body := map[string]interface{}{
		"sync_product": map[string]interface{}{
			"name":      datas.Name,
			"thumbnail": thumbnailURL,
		},
		"sync_variants": syncVariants,
	}

	log.Println(body)*/
	//body := map[string]interface{}{}
	/*
			err := mapstructure.Decode(request.Order, &body)
			if err != nil {
				log.Println(err)
				return nil, errors.New("error while decoding request")
			}
		* /

		log.Println(body)

		headers := map[string]string{
			"Authorization": "Bearer " + printfulConfig.AccessToken,
		}

		resp, err := fetchRateLimited("POST", PRINTFUL_ORDERS_API, "", headers, body)
		if err != nil {
			return nil, errors.New("unable to get printful response")
		}

		//body2, _ := ioutil.ReadAll(resp.Body)
		//log.Println(string(body2))

		response := CreateOrderResponse{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Println(err)
			return nil, errors.New("unable to decode printful response")
		}

		log.Println(response)

		p := &(response.Result)

		return p, nil
	*/
}
