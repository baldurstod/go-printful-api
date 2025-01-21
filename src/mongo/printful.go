package mongo

import (
	"context"
	"go-printful-api/src/config"
	"log"
	"time"

	model "github.com/baldurstod/printful-api-model"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cancelConnect context.CancelFunc
var productsCollection *mongo.Collection
var variantsCollection *mongo.Collection

var cacheMaxAge int64 = 86400

func InitPrintfulDB(config config.Database) {
	log.Println(config)
	var ctx context.Context
	ctx, cancelConnect = context.WithCancel(context.Background())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.ConnectURI))
	if err != nil {
		log.Println(err)
		panic(err)
	}

	defer closePrintfulDB()

	productsCollection = client.Database(config.DBName).Collection("products")
	variantsCollection = client.Database(config.DBName).Collection("variants")
}

func closePrintfulDB() {
	if cancelConnect != nil {
		cancelConnect()
	}
}

type MongoProductInfo struct {
	ID          int               `json:"id" bson:"id"`
	LastUpdated int64             `json:"last_updated" bson:"last_updated"`
	ProductInfo model.ProductInfo `json:"product_info" bson:"product_info"`
}

func FindProduct(productID int) (*model.ProductInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"id", productID}}

	r := productsCollection.FindOne(ctx, filter)

	doc := MongoProductInfo{}
	if err := r.Decode(&doc); err != nil {
		return nil, err
	}

	if time.Now().Unix()-doc.LastUpdated > cacheMaxAge {
		return &doc.ProductInfo, MaxAgeError{}
	}

	return &doc.ProductInfo, nil
}

func InsertProduct(productInfo *model.ProductInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	filter := bson.D{{"id", productInfo.Product.ID}}
	doc := MongoProductInfo{ID: productInfo.Product.ID, LastUpdated: time.Now().Unix(), ProductInfo: *productInfo}
	_, err := productsCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}

type MongoVariantInfo struct {
	ID          int               `json:"id" bson:"id"`
	LastUpdated int64             `json:"last_updated" bson:"last_updated"`
	VariantInfo model.VariantInfo `json:"variant_info" bson:"variant_info"`
}

func FindVariant(variantID int) (*model.VariantInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"id", variantID}}

	r := variantsCollection.FindOne(ctx, filter)

	doc := MongoVariantInfo{}
	if err := r.Decode(&doc); err != nil {
		return nil, err
	}

	if time.Now().Unix()-doc.LastUpdated > cacheMaxAge {
		return &doc.VariantInfo, MaxAgeError{}
	}

	return &doc.VariantInfo, nil
}

func InsertVariant(variantInfo *model.VariantInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	filter := bson.D{{"id", variantInfo.Variant.ID}}
	doc := MongoVariantInfo{ID: variantInfo.Variant.ID, LastUpdated: time.Now().Unix(), VariantInfo: *variantInfo}
	_, err := variantsCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}
