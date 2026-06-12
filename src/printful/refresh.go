package printful

import (
	"fmt"
	"go-printful-api/src/database"
	"log"

	printfulsdk "github.com/baldurstod/go-printful-sdk"
)

func RefreshCountries() error {
	countries, err := printfulClient.GetCountries()

	if err != nil {
		return fmt.Errorf("error in RefreshCountries while fetching countries: %w", err)
	}

	for _, country := range countries {
		err = database.InsertCountry(&country)
		if err != nil {
			log.Println("error in RefreshCountries:", err)
		}
	}
	return nil
}

func RefreshCategories(language string) error {
	categories, err := printfulClient.GetCatalogCategories(printfulsdk.WithLanguage(language))

	if err != nil {
		return fmt.Errorf("error in RefreshCategories while fetching categories: %w", err)
	}

	for _, category := range categories {
		err = database.InsertCategory(&category, language)
		if err != nil {
			log.Println("error in RefreshCategories:", err)
		}
	}
	return nil
}
