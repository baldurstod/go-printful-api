package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
	"github.com/lib/pq"
)

func InsertProduct(product *printfulmodel.Product) error {
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to init postgre ?")
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

	_, err = printfulDb.Exec(`INSERT INTO products (id, main_category_id, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, last_updated)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	ON CONFLICT (id) DO UPDATE SET
	main_category_id = $2,
	type = $3,
	name = $4,
	brand = $5,
	model = $6,
	image = $7,
	variant_count = $8,
	catalog_variant_ids = $9,
	is_discontinued = $10,
	description = $11,
	sizes = $12,
	colors = $13,
	techniques = $14,
	placements = $15,
	product_options = $16,
	last_updated = $17`,
		product.ID,
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

func FindProducts() ([]printfulmodel.Product, error) {
	if printfulDb == nil {
		return nil, errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	query := `SELECT id, main_category_id, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, last_updated FROM products;`
	res, err := printfulDb.Query(query)
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
		var catalogVariantIDs []int32
		var isDiscontinued bool
		var description string
		var sizes []string
		var colors string
		var techniques string
		var placements string
		var productOptions string
		var lastUpdated int64

		err = res.Scan(&id, &mainCategoryID, &productType, &name, &brand, &model, &image, &variantCount, pq.Array(&catalogVariantIDs), &isDiscontinued, &description, pq.Array(&sizes), &colors, &techniques, &placements, &productOptions, &lastUpdated)
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

		catalogVariantIDs2 := make([]int, len(catalogVariantIDs))
		for i, i32 := range catalogVariantIDs {
			catalogVariantIDs2[i] = int(i32)
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
			CatalogVariantIDs: catalogVariantIDs2,
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

func FindProduct(productID int) (*printfulmodel.Product, bool, error) {
	if printfulDb == nil {
		return nil, false, errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	query := `SELECT id, main_category_id, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, last_updated FROM products WHERE id = $1;`
	row := printfulDb.QueryRow(query, productID)

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
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	_, err := printfulDb.Exec(`UPDATE products SET variant_count = $1, catalog_variant_ids = $2 WHERE id = $3`,
		len(variantIds),
		variantIds,
		productID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product "+strconv.Itoa(productID)+" : <%w>", err)
	}

	return nil
}
