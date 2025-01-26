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
