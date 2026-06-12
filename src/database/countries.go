package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
)

func InsertCountry(country *printfulmodel.Country) error {
	if db == nil {
		return errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	j, err := json.Marshal(&country.States)
	if err != nil {
		return fmt.Errorf("failed to marshal country.States: <%w>", err)
	}

	_, err = db.Exec(`INSERT INTO countries (code, name, region, states, last_updated)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (code) DO UPDATE SET
	name = $2,
	region = $3,
	states = $4,
	last_updated = $5`,
		country.Code,
		country.Name,
		country.Region,
		j,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert country "+country.Code+" : <%w>", err)
	}

	return nil
}
