package mongo

import (
	"context"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCategory struct {
	ID          int                    `json:"id" bson:"id"`
	LastUpdated int64                  `json:"last_updated" bson:"last_updated"`
	Category    printfulmodel.Category `json:"category" bson:"category"`
}

func InsertCategory(category *printfulmodel.Category) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5005*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	filter := bson.D{{Key: "id", Value: category.ID}}
	doc := MongoCategory{ID: category.ID, LastUpdated: time.Now().Unix(), Category: *category}
	_, err := categoriesCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}

func FindCategories() ([]printfulmodel.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{}

	cursor, err := categoriesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	categories := make([]printfulmodel.Category, 0, 400)
	for cursor.Next(context.TODO()) {
		doc := MongoCategory{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		categories = append(categories, doc.Category)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
