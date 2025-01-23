package mongo

import (
	"context"
	"go-printful-api/src/config"
	"log"
	"time"

	printfulmodel "github.com/baldurstod/go-printful-sdk/model"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cancelConnect context.CancelFunc
var productsCollection *mongo.Collection
var productsPricesCollection *mongo.Collection
var productsTemplatesCollection *mongo.Collection
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
	productsPricesCollection = client.Database(config.DBName).Collection("products_prices")
	productsTemplatesCollection = client.Database(config.DBName).Collection("products_templates")
	variantsCollection = client.Database(config.DBName).Collection("variants")
}

func closePrintfulDB() {
	if cancelConnect != nil {
		cancelConnect()
	}
}

type MongoProduct struct {
	ID          int                   `json:"id" bson:"id"`
	LastUpdated int64                 `json:"last_updated" bson:"last_updated"`
	Product     printfulmodel.Product `json:"product" bson:"product"`
}

func FindProducts() ([]printfulmodel.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{}

	cursor, err := productsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	products := make([]printfulmodel.Product, 0, 400)
	for cursor.Next(context.TODO()) {
		doc := MongoProduct{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		products = append(products, doc.Product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func FindProduct(productID int) (*printfulmodel.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "id", Value: productID}}

	r := productsCollection.FindOne(ctx, filter)

	doc := MongoProduct{}
	if err := r.Decode(&doc); err != nil {
		return nil, err
	}

	if time.Now().Unix()-doc.LastUpdated > cacheMaxAge {
		return &doc.Product, MaxAgeError{}
	}

	return &doc.Product, nil
}

func FindVariants(productID int) (variants []printfulmodel.Variant, outdated bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "variant.catalog_product_id", Value: productID}}
	outdated = false

	cursor, err := variantsCollection.Find(ctx, filter)
	if err != nil {
		return nil, false, err
	}

	variants = make([]printfulmodel.Variant, 0, 20)
	for cursor.Next(context.TODO()) {
		doc := MongoVariant{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, false, err
		}

		if time.Now().Unix()-doc.LastUpdated > cacheMaxAge {
			outdated = true
		}

		variants = append(variants, doc.Variant)
	}

	if err := cursor.Err(); err != nil {
		return nil, false, err
	}

	return variants, outdated, nil
}

func InsertProduct(product *printfulmodel.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5005*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	filter := bson.D{{Key: "id", Value: product.ID}}
	doc := MongoProduct{ID: product.ID, LastUpdated: time.Now().Unix(), Product: *product}
	_, err := productsCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}

type MongoProductPrices struct {
	ProductID     int                         `json:"product_id" bson:"product_id"`
	Currency      string                      `json:"currency" bson:"currency"`
	LastUpdated   int64                       `json:"last_updated" bson:"last_updated"`
	ProductPrices printfulmodel.ProductPrices `json:"product_prices" bson:"product_prices"`
}

func InsertProductPrices(productPrices *printfulmodel.ProductPrices) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	//	filter := bson.D{{Key: "id", Value: productPrices.Product.ID}}
	filter := bson.D{
		{Key: "$and",
			Value: bson.A{
				bson.D{{Key: "product_id", Value: productPrices.Product.ID}},
				bson.D{{Key: "currency", Value: productPrices.Currency}},
			},
		},
	}

	doc := MongoProductPrices{ProductID: productPrices.Product.ID, Currency: productPrices.Currency, LastUpdated: time.Now().Unix(), ProductPrices: *productPrices}
	_, err := productsPricesCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}

type MongoProductTemplates struct {
	ProductID        int                           `json:"product_id" bson:"product_id"`
	LastUpdated      int64                         `json:"last_updated" bson:"last_updated"`
	ProductTemplates printfulmodel.ProductTemplate `json:"product_templates" bson:"product_templates"`
}

func InsertProductTemplates(productID int, productTemplates *printfulmodel.ProductTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	filter := bson.D{{Key: "product_id", Value: productID}}

	doc := MongoProductTemplates{ProductID: productID, LastUpdated: time.Now().Unix(), ProductTemplates: *productTemplates}
	_, err := productsTemplatesCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}

type MongoVariant struct {
	ID          int                   `json:"id" bson:"id"`
	LastUpdated int64                 `json:"last_updated" bson:"last_updated"`
	Variant     printfulmodel.Variant `json:"variant" bson:"variant"`
}

func FindVariant(variantID int) (*printfulmodel.Variant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "id", Value: variantID}}

	r := variantsCollection.FindOne(ctx, filter)

	doc := MongoVariant{}
	if err := r.Decode(&doc); err != nil {
		return nil, err
	}

	if time.Now().Unix()-doc.LastUpdated > cacheMaxAge {
		return &doc.Variant, MaxAgeError{}
	}

	return &doc.Variant, nil
}

func InsertVariant(variant *printfulmodel.Variant) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)

	filter := bson.D{{Key: "id", Value: variant.ID}}
	doc := MongoVariant{ID: variant.ID, LastUpdated: time.Now().Unix(), Variant: *variant}
	_, err := variantsCollection.ReplaceOne(ctx, filter, doc, opts)

	return err
}
