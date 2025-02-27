package main_test

import (
	"encoding/json"
	"go-printful-api/src/config"
	"go-printful-api/src/mongo"
	"go-printful-api/src/printful"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"testing"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_, filename, _, _ := runtime.Caller(0)
	// The ".." may change depending on you folder structure
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	err = initConfig()
	if err != nil {
		panic(err)
	}
}

func RefreshAllProducts() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		printful.RefreshAllProducts("USD", true)
	}()
	wg.Wait()
}

func initConfig() error {
	var err error
	var content []byte
	config := config.Config{}

	if content, err = os.ReadFile("config.json"); err != nil {
		return err
	}
	if err = json.Unmarshal(content, &config); err != nil {
		return err
	}
	printful.SetPrintfulConfig(config.Printful)
	mongo.InitPrintfulDB(config.Databases.Printful)
	mongo.InitImagesDB(config.Databases.Images)
	return nil
}

func TestGetProducts(t *testing.T) {
	products, err := printful.GetProducts()
	if err != nil {
		t.Error(err)
		return
	}

	j, _ := json.MarshalIndent(&products, "", "")

	err = os.WriteFile("./var/products.json", j, 0666)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetProduct(t *testing.T) {
	products, err := printful.GetProduct(638)
	if err != nil {
		t.Error(err)
		return
	}

	j, _ := json.MarshalIndent(&products, "", "")

	err = os.WriteFile("./var/products_638.json", j, 0666)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetVariants(t *testing.T) {
	products, err := printful.GetVariants(679)
	if err != nil {
		t.Error(err)
		return
	}

	j, _ := json.MarshalIndent(&products, "", "")

	err = os.WriteFile("./var/product_variants.json", j, 0666)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestTemplates(t *testing.T) {
	variants, err := printful.GetVariants(679)
	if err != nil {
		t.Error(err)
		return
	}

	j, _ := json.MarshalIndent(&variants, "", "")

	err = os.WriteFile("./var/product_variants.json", j, 0666)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestRefreshAllProducts(t *testing.T) {
	RefreshAllProducts()
}

func TestRefreshCountries(t *testing.T) {
	printful.RefreshCountries()
}

func TestRefreshCategories(t *testing.T) {
	printful.RefreshCategories()
}

func TestTemplatesWithMultipleTechniques(t *testing.T) {
	products, err := printful.GetProducts()
	if err != nil {
		t.Error(err)
		return
	}

	techniques := make(map[int]string)
	variants := make(map[int]int)
	count := 0

	for _, product := range products {
		templates, _, err := mongo.FindMockupTemplates(product.ID)
		if err != nil {
			log.Println("error while finding templates for product", product.ID, err)
		}

		for templateId, template := range templates {
			for _, variantId := range template.CatalogVariantIDs {
				technique, found := techniques[variantId]
				if found && technique != template.Technique {
					log.Println("different techniques for variant", variantId, "template", templateId, "product", product.ID)
					variants[variantId] = variantId
					count++
				} else {
					techniques[variantId] = template.Technique
				}
			}
		}
	}
	log.Println("total", count, len(variants))

}

func TestProductPrices(t *testing.T) {
	prices, err := printful.GetProductPrices(679, "USD")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("prices", prices)

}

func TestGetSimilarVariants(t *testing.T) {
	testCases := make(map[int][]printful.GetSimilarVariantsPlacement)

	testCases[19662] = []printful.GetSimilarVariantsPlacement{
		{Placement: "chest_left_dtf", Technique: "dtfilm", Orientation: "any"},
	}

	testCases[20251] = []printful.GetSimilarVariantsPlacement{
		{Placement: "default", Technique: "uv", Orientation: "any"},
	}

	for variantID, placements := range testCases {
		variants, err := printful.GetSimilarVariants(variantID, placements)
		if err != nil {
			t.Error(err)
		}
		log.Printf("similar variants for %d %v", variantID, variants)
	}

}
