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

	_, err = db.Exec(`INSERT INTO products (id, language, main_category_id, type, name, brand, model, image, variant_count, is_discontinued, description, sizes, colors, techniques, placements, last_updated)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (id, language) DO UPDATE SET
	main_category_id = $3,
	type = $4,
	name = $5,
	brand = $6,
	model = $7,
	image = $8,
	variant_count = $9,
	is_discontinued = $10,
	description = $11,
	sizes = $12,
	colors = $13,
	techniques = $14,
	placements = $15,
	last_updated = $16`,
		product.ID,
		language,
		product.MainCategoryID,
		product.Type,
		product.Name,
		product.Brand,
		product.Model,
		product.Image,
		product.VariantCount,
		product.IsDiscontinued,
		product.Description,
		product.Sizes,
		colors,
		techniques,
		placements,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert product "+strconv.Itoa(product.ID)+" : <%w>", err)
	}

	return nil
}
