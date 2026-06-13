package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
)

func InsertMockupStyles(productID int, mockupStyles []printfulmodel.MockupStyles) error {
	if db == nil {
		return errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	styles, err := json.Marshal(&mockupStyles)
	if err != nil {
		return fmt.Errorf("failed to marshal mockupStyles: <%w>", err)
	}

	_, err = db.Exec(`INSERT INTO mockup_styles (product_id, mockup_styles, last_updated)
	VALUES ($1, $2, $3)
	ON CONFLICT (product_id) DO UPDATE SET
	mockup_styles = $2,
	last_updated = $3`,
		productID,
		styles,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert mockup styles "+strconv.Itoa(productID)+" : <%w>", err)
	}

	return nil
}

func FindMockupStyles(productID int) ([]printfulmodel.MockupStyles, bool, error) {

	if db == nil {
		return nil, false, errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	query := `SELECT mockup_styles, last_updated FROM mockup_styles WHERE product_id = $1;`
	row := db.QueryRow(query, productID)

	var mockupStyles string
	var lastUpdated int64

	err := row.Scan(&mockupStyles, &lastUpdated)
	if err != nil {
		return nil, false, fmt.Errorf("failed to scan row in FindMockupStyles: <%w>", err)
	}

	styles := []printfulmodel.MockupStyles{}
	if err = json.Unmarshal([]byte(mockupStyles), &styles); err != nil {
		return nil, false, err
	}

	return styles, time.Now().Unix()-lastUpdated > cacheMaxAge, nil
}
