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
