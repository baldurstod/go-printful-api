package mongo

import (
	"context"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCountry struct {
	Code        string                `json:"code" bson:"code"`
	LastUpdated int64                 `json:"last_updated" bson:"last_updated"`
	Country     printfulmodel.Country `json:"country" bson:"country"`
}

func InsertCountry(country *printfulmodel.Country) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5005*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	filter := bson.D{{Key: "code", Value: country.Code}}
	doc := MongoCountry{Code: country.Code, LastUpdated: time.Now().Unix(), Country: *country}
	_, err := countriesCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}

func FindCountries() ([]printfulmodel.Country, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{}

	cursor, err := countriesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	countries := make([]printfulmodel.Country, 0, 400)
	for cursor.Next(context.TODO()) {
		doc := MongoCountry{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		countries = append(countries, doc.Country)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}
