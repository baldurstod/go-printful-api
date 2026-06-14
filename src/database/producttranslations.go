package database

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
)

type ProductTranslation struct {
	ID          int    `json:"id" bson:"id" mapstructure:"id"`
	Language    string `json:"language" bson:"language" mapstructure:"language"`
	Name        string `json:"name" bson:"name" mapstructure:"name"`
	Description string `json:"description" bson:"description" mapstructure:"description"`
	LastUpdated int64  `json:"last_updated" bson:"last_updated"`
}

func InsertProductTranslation(language string, product *printfulmodel.Product) error {
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	_, err := printfulDb.Exec(`INSERT INTO product_translations (id, language, name, description, last_updated)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id, language) DO UPDATE SET
	name = $3,
	description = $4,
	last_updated = $5`,
		product.ID,
		language,
		product.Name,
		product.Description,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert product translation"+strconv.Itoa(product.ID)+" "+language+" : <%w>", err)
	}

	return nil
}

func FindProductTranslations(language string) ([]ProductTranslation, error) {
	if printfulDb == nil {
		return nil, errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	query := `SELECT product_id, name, description, last_updated FROM product_translations WHERE language = $1;`
	res, err := printfulDb.Query(query, language)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query "+query+"in FindProductTranslations: <%w>", err)
	}
	defer res.Close()

	productTranslations := make([]ProductTranslation, 0, 400)
	for res.Next() {
		var id int
		var name string
		var description string
		var lastUpdated int64

		err = res.Scan(&name, &description, &lastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row in FindProductTranslations: <%w>", err)
		}

		productTranslation := ProductTranslation{
			ID:          id,
			Language:    language,
			Name:        name,
			Description: description,
			LastUpdated: lastUpdated,
		}

		productTranslations = append(productTranslations, productTranslation)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to get next row in FindProductTranslations: <%w>", err)
	}

	return productTranslations, nil
}

func FindProductTranslation(productID int, language string) (*ProductTranslation, error) {
	if printfulDb == nil {
		return nil, errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	query := `SELECT name, description, last_updated FROM product_translations WHERE product_id = $1 AND language = $2;`
	row := printfulDb.QueryRow(query, productID, language)

	var id int
	var name string
	var description string
	var lastUpdated int64

	err := row.Scan(&name, &description, &lastUpdated)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row in FindProductTranslation: <%w>", err)
	}

	productTranslation := ProductTranslation{
		ID:          id,
		Language:    language,
		Name:        name,
		Description: description,
		LastUpdated: lastUpdated,
	}

	return &productTranslation, nil
}
