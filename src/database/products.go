package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
)

func InsertProduct(product *printfulmodel.Product, language string) error {
	if db == nil {
		return errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	colors, err := json.Marshal(&product.Colors)
	if err != nil {
		return fmt.Errorf("failed to marshal product.Colors: <%w>", err)
	}

	techniques, err := json.Marshal(&product.Techniques)
	if err != nil {
		return fmt.Errorf("failed to marshal product.Techniques: <%w>", err)
	}

	placements, err := json.Marshal(&product.Placements)
	if err != nil {
		return fmt.Errorf("failed to marshal product.Placements: <%w>", err)
	}

	productOptions, err := json.Marshal(&product.ProductOptions)
	if err != nil {
		return fmt.Errorf("failed to marshal product.ProductOptions: <%w>", err)
	}

	catalogVariantIDs := product.CatalogVariantIDs
	if catalogVariantIDs == nil {
		catalogVariantIDs = []int{}
	}

	_, err = db.Exec(`INSERT INTO products (id, language, main_category_id, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, last_updated)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	ON CONFLICT (id, language) DO UPDATE SET
	main_category_id = $3,
	type = $4,
	name = $5,
	brand = $6,
	model = $7,
	image = $8,
	variant_count = $9,
	catalog_variant_ids = $10,
	is_discontinued = $11,
	description = $12,
	sizes = $13,
	colors = $14,
	techniques = $15,
	placements = $16,
	product_options = $17,
	last_updated = $18`,
		product.ID,
		language,
		product.MainCategoryID,
		product.Type,
		product.Name,
		product.Brand,
		product.Model,
		product.Image,
		product.VariantCount,
		catalogVariantIDs,
		product.IsDiscontinued,
		product.Description,
		product.Sizes,
		colors,
		techniques,
		placements,
		productOptions,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert product "+strconv.Itoa(product.ID)+" : <%w>", err)
	}

	return nil
}

func FindProducts(language string) ([]printfulmodel.Product, error) {
	if db == nil {
		return nil, errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	query := `SELECT id, main_category_id, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, last_updated FROM products WHERE language = $1;`
	res, err := db.Query(query, language)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query "+query+"in FindProducts: <%w>", err)
	}
	defer res.Close()

	products := make([]printfulmodel.Product, 0, 400)
	for res.Next() {
		var id int
		var mainCategoryID int
		var productType string
		var name string
		var brand string
		var model string
		var image string
		var variantCount int
		var catalogVariantIDs []int
		var isDiscontinued bool
		var description string
		var sizes []string
		var colors string
		var techniques string
		var placements string
		var productOptions string
		var lastUpdated int64

		err = res.Scan(&id, &mainCategoryID, &productType, &name, &brand, &model, &image, &variantCount, &catalogVariantIDs, &isDiscontinued, &description, &sizes, &colors, &techniques, &placements, &productOptions, &lastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row in FindProducts: <%w>", err)
		}

		jsonColors := []printfulmodel.Color{}
		if err = json.Unmarshal([]byte(colors), &jsonColors); err != nil {
			return nil, err
		}

		jsonTechniques := []printfulmodel.Technique{}
		if err = json.Unmarshal([]byte(techniques), &jsonTechniques); err != nil {
			return nil, err
		}

		jsonPlacements := []printfulmodel.ProductPlacement{}
		if err = json.Unmarshal([]byte(placements), &jsonPlacements); err != nil {
			return nil, err
		}

		jsonProductOptions := []printfulmodel.CatalogOption{}
		if err = json.Unmarshal([]byte(productOptions), &jsonProductOptions); err != nil {
			return nil, err
		}

		product := printfulmodel.Product{
			ID:                id,
			MainCategoryID:    mainCategoryID,
			Type:              productType,
			Name:              name,
			Brand:             brand,
			Model:             model,
			Image:             image,
			VariantCount:      variantCount,
			CatalogVariantIDs: catalogVariantIDs,
			IsDiscontinued:    isDiscontinued,
			Description:       description,
			Sizes:             sizes,
			Colors:            jsonColors,
			Techniques:        jsonTechniques,
			Placements:        jsonPlacements,
			ProductOptions:    jsonProductOptions,
		}

		products = append(products, product)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to get next row in FindProducts: <%w>", err)
	}

	return products, nil
}

func FindProduct(productID int, language string) (*printfulmodel.Product, bool, error) {
	if db == nil {
		return nil, false, errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	query := `SELECT id, main_category_id, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, last_updated FROM products WHERE id = $1 AND language = $2;`
	row := db.QueryRow(query, productID, language)

	var id int
	var mainCategoryID int
	var productType string
	var name string
	var brand string
	var model string
	var image string
	var variantCount int
	var catalogVariantIDs []int
	var isDiscontinued bool
	var description string
	var sizes []string
	var colors string
	var techniques string
	var placements string
	var productOptions string
	var lastUpdated int64

	err := row.Scan(&id, &mainCategoryID, &productType, &name, &brand, &model, &image, &variantCount, &catalogVariantIDs, &isDiscontinued, &description, &sizes, &colors, &techniques, &placements, &productOptions, &lastUpdated)
	if err != nil {
		return nil, false, fmt.Errorf("failed to scan row in FindProduct: <%w>", err)
	}

	jsonColors := []printfulmodel.Color{}
	if err = json.Unmarshal([]byte(colors), &jsonColors); err != nil {
		return nil, false, err
	}

	jsonTechniques := []printfulmodel.Technique{}
	if err = json.Unmarshal([]byte(techniques), &jsonTechniques); err != nil {
		return nil, false, err
	}

	jsonPlacements := []printfulmodel.ProductPlacement{}
	if err = json.Unmarshal([]byte(placements), &jsonPlacements); err != nil {
		return nil, false, err
	}

	jsonProductOptions := []printfulmodel.CatalogOption{}
	if err = json.Unmarshal([]byte(productOptions), &jsonProductOptions); err != nil {
		return nil, false, err
	}

	product := printfulmodel.Product{
		ID:                id,
		MainCategoryID:    mainCategoryID,
		Type:              productType,
		Name:              name,
		Brand:             brand,
		Model:             model,
		Image:             image,
		VariantCount:      variantCount,
		CatalogVariantIDs: catalogVariantIDs,
		IsDiscontinued:    isDiscontinued,
		Description:       description,
		Sizes:             sizes,
		Colors:            jsonColors,
		Techniques:        jsonTechniques,
		Placements:        jsonPlacements,
		ProductOptions:    jsonProductOptions,
	}

	return &product, time.Now().Unix()-lastUpdated > cacheMaxAge, nil
}

func UpdateProductVariantIds(productID int, variantIds []int) error {
	if db == nil {
		return errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	_, err := db.Exec(`UPDATE products SET variant_count = $1, catalog_variant_ids = $2 WHERE id = $3`,
		len(variantIds),
		variantIds,
		productID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product "+strconv.Itoa(productID)+" : <%w>", err)
	}

	return nil
}
