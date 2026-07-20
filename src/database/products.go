package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
	"github.com/lib/pq"
)

func InsertProduct(product printfulmodel.Product) error {
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	if product.Categories == nil {
		product.Categories = make([]int, 0)
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

	_, err = printfulDb.Exec(`INSERT INTO products (id, main_category_id, categories, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, date_created, date_updated)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	ON CONFLICT (id) DO UPDATE SET
	main_category_id = $2,
	categories = $3,
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
	date_created = $18,
	date_updated = $19`,
		product.ID,
		product.MainCategoryID,
		product.Categories,
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
		time.Now(),
		time.Now(),
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

	query := `SELECT id, main_category_id, type, name, brand, model, image, variant_count, catalog_variant_ids, is_discontinued, description, sizes, colors, techniques, placements, product_options, date_created, date_updated FROM products;`
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
		var dateCreated time.Time
		var dateUpdated time.Time

		err = res.Scan(&id, &mainCategoryID, &productType, &name, &brand, &model, &image, &variantCount, pq.Array(&catalogVariantIDs), &isDiscontinued, &description, pq.Array(&sizes), &colors, &techniques, &placements, &productOptions, &dateCreated, &dateUpdated)
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

	err := row.Scan(&id, &mainCategoryID, &productType, &name, &brand, &model, &image, &variantCount, pq.Array(&catalogVariantIDs), &isDiscontinued, &description, pq.Array(&sizes), &colors, &techniques, &placements, &productOptions, &lastUpdated)
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

type UpdateProductFields struct {
	MainCategoryID    bool
	Categories        bool
	Type              bool
	Name              bool
	Brand             bool
	Model             bool
	Image             bool
	ImageWomen        bool
	VariantCount      bool
	CatalogVariantIDs bool
	IsDiscontinued    bool
	Description       bool
	Sizes             bool
	Colors            bool
	Techniques        bool
	Placements        bool
	ProductOptions    bool
}

func UpdateProduct(product printfulmodel.Product, fields UpdateProductFields) error {
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	queryString := make([]string, 0)
	queryParams := []any{product.ID, time.Now()}

	v := reflect.ValueOf(fields)
	typeOfS := v.Type()

	// Using reflection to list UpdateProductFields fields
	for i := 0; i < v.NumField(); i++ {
		name := typeOfS.Field(i).Name
		value := v.Field(i).Bool()
		if !value {
			continue
		}

		addSetStatement := func(column string, value any) {
			param := "$" + strconv.Itoa(len(queryParams)+1)

			queryString = append(queryString, column+" = "+param)
			queryParams = append(queryParams, value)
		}

		switch name {
		case "MainCategoryID":
			addSetStatement("main_category_id", product.MainCategoryID)
		case "Categories":
			addSetStatement("categories", product.Categories)
		case "Type":
			addSetStatement("type", product.Type)
		case "Name":
			addSetStatement("name", product.Name)
		case "Brand":
			addSetStatement("brand", product.Brand)
		case "Model":
			addSetStatement("model", product.Model)
		case "Image":
			addSetStatement("image", product.Image)
		case "ImageWomen":
			addSetStatement("image_women", product.ImageWomen)
		case "VariantCount":
			addSetStatement("variant_count", product.VariantCount)
		case "CatalogVariantIDs":
			addSetStatement("catalog_variant_ids", product.CatalogVariantIDs)
		case "IsDiscontinued":
			addSetStatement("is_discontinued", product.IsDiscontinued)
		case "Description":
			addSetStatement("description", product.Description)
		case "Sizes":
			addSetStatement("sizes", product.Sizes)
		case "Colors":
			addSetStatement("colors", product.Colors)
		case "Techniques":
			addSetStatement("techniques", product.Techniques)
		case "Placements":
			addSetStatement("placements", product.Placements)
		case "ProductOptions":
			addSetStatement("product_options", product.ProductOptions)
		default:
			return errors.New("missing field in UpdateProduct " + name)
		}
	}

	if len(queryString) == 0 {
		return errors.New("failed to update product: no field selected for update")
	}

	query := `UPDATE products SET date_updated = $2,` + strings.Join(queryString, ",") + ` WHERE id = $1;`
	_, err := printfulDb.Exec(query, queryParams...)

	if err != nil {
		return fmt.Errorf("failed to update product: <%w>", err)
	}

	return nil
}
