package database

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
)

func InsertCategory(category *printfulmodel.Category, language string) error {
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	_, err := printfulDb.Exec(`INSERT INTO categories (id, language, parent_id, image_url, title, last_updated)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (id, language) DO UPDATE SET
	parent_id = $3,
	image_url = $4,
	title = $5,
	last_updated = $6`,
		category.ID,
		language,
		category.ParentID,
		category.ImageURL,
		category.Title,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert category "+strconv.Itoa(category.ID)+" : <%w>", err)
	}

	return nil
}

func FindCategories() ([]printfulmodel.Category, error) {
	if printfulDb == nil {
		return nil, errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	query := `SELECT id, parent_id, image_url, title FROM categories;`
	res, err := printfulDb.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query "+query+"in FindCategories: <%w>", err)
	}
	defer res.Close()

	categories := make([]printfulmodel.Category, 0, 400)
	for res.Next() {
		var id int
		var parent_id int
		var image_url string
		var title string

		err = res.Scan(&id, &parent_id, &image_url, &title)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row in FindCategories: <%w>", err)
		}
		category := printfulmodel.Category{ID: id, ParentID: parent_id, ImageURL: image_url, Title: title}

		categories = append(categories, category)
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to get next row in FindCategories: <%w>", err)
	}

	return categories, nil
}
