package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
)

func InsertMockupTemplates(productID int, mockupTemplates []printfulmodel.MockupTemplates) error {
	if db == nil {
		return errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	templates, err := json.Marshal(&mockupTemplates)
	if err != nil {
		return fmt.Errorf("failed to marshal mockupTemplates: <%w>", err)
	}

	_, err = db.Exec(`INSERT INTO mockup_templates (product_id, mockup_templates, last_updated)
	VALUES ($1, $2, $3)
	ON CONFLICT (product_id) DO UPDATE SET
	mockup_templates = $2,
	last_updated = $3`,
		productID,
		templates,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert mockup templates "+strconv.Itoa(productID)+" : <%w>", err)
	}

	return nil
}

func FindMockupTemplates(productID int) ([]printfulmodel.MockupTemplates, bool, error) {
	if db == nil {
		return nil, false, errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	query := `SELECT mockup_templates, last_updated FROM mockup_templates WHERE product_id = $1;`
	row := db.QueryRow(query, productID)

	var mockupTemplates string
	var lastUpdated int64

	err := row.Scan(&mockupTemplates, &lastUpdated)
	if err != nil {
		return nil, false, fmt.Errorf("failed to scan row in FindMockupTemplates: <%w>", err)
	}

	templates := []printfulmodel.MockupTemplates{}
	if err = json.Unmarshal([]byte(mockupTemplates), &templates); err != nil {
		return nil, false, err
	}

	return templates, time.Now().Unix()-lastUpdated > cacheMaxAge, nil
}
