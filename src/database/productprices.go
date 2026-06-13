package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
)

func InsertProductPrices(productPrices *printfulmodel.ProductPrices) error {
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	prices, err := json.Marshal(&productPrices)
	if err != nil {
		return fmt.Errorf("failed to marshal productPrices: <%w>", err)
	}

	_, err = printfulDb.Exec(`INSERT INTO products_prices (product_id, currency, product_prices, last_updated)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (product_id, currency) DO UPDATE SET
	product_prices = $3,
	last_updated = $4`,
		productPrices.Product.ID,
		productPrices.Currency,
		prices,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert product prices "+strconv.Itoa(productPrices.Product.ID)+" "+productPrices.Currency+" : <%w>", err)
	}

	return nil
}

func FindProductPrices(productID int, currency string) (*printfulmodel.ProductPrices, bool, error) {
	if printfulDb == nil {
		return nil, false, errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	query := `SELECT product_prices, last_updated FROM products_prices WHERE product_id = $1 AND currency = $2;`
	row := printfulDb.QueryRow(query, productID, currency)

	var productPrices string
	var lastUpdated int64

	err := row.Scan(&productPrices, &lastUpdated)
	if err != nil {
		return nil, false, fmt.Errorf("failed to scan row in FindProductPrices: <%w>", err)
	}

	prices := printfulmodel.ProductPrices{}
	if err = json.Unmarshal([]byte(productPrices), &prices); err != nil {
		return nil, false, err
	}

	return &prices, time.Now().Unix()-lastUpdated > cacheMaxAge, nil
}
