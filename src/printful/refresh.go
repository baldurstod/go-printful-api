package printful

import (
	"fmt"
	"go-printful-api/src/mongo"
)

func RefreshCountries() error {
	countries, err := printfulClient.GetCountries()

	if err != nil {
		return fmt.Errorf("error in RefreshCountries while fetching countries: %w", err)
	}

	for _, country := range countries {
		mongo.InsertCountry(&country)
	}
	return nil
}

func RefreshCategories() error {
	categories, err := printfulClient.GetCatalogCategories()

	if err != nil {
		return fmt.Errorf("error in RefreshCategories while fetching categories: %w", err)
	}

	for _, category := range categories {
		mongo.InsertCategory(&category)
	}
	return nil
}
