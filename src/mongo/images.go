package mongo

import (
	"bytes"
	"context"
	"go-printful-api/src/config"
	"image"
	"image/png"
	"log"
	_ "time"

	_ "github.com/baldurstod/printful-api-model"
	_ "go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cancelImagesConnect context.CancelFunc
var imagesBucket *gridfs.Bucket

func InitImagesDB(config config.Database) {
	log.Println(config)
	var ctx context.Context
	ctx, cancelImagesConnect = context.WithCancel(context.Background())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.ConnectURI))
	if err != nil {
		log.Println(err)
		panic(err)
	}

	defer closeImagesDB()

	imagesBucket, err = gridfs.NewBucket(client.Database(config.DBName), options.GridFSBucket().SetName(config.BucketName))
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func closeImagesDB() {
	if cancelImagesConnect != nil {
		cancelImagesConnect()
	}
}

func UploadImage(filename string, img image.Image) error {
	uploadStream, err := imagesBucket.OpenUploadStream(filename)
	if err != nil {
		return err
	}

	defer uploadStream.Close()

	buf := bytes.Buffer{}
	err = png.Encode(&buf, img)
	if err != nil {
		return err
	}

	//log.Println(buf)
	fileSize, err := uploadStream.Write(buf.Bytes())
	log.Println(fileSize, err)

	return nil
}
