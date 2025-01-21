package main

import (
	"encoding/json"
	"go-printful-api/src/config"
	"go-printful-api/src/mongo"
	"go-printful-api/src/printful"
	"go-printful-api/src/server"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	config := config.Config{}
	if content, err := os.ReadFile("config.json"); err == nil {
		if err = json.Unmarshal(content, &config); err == nil {
			printful.SetPrintfulConfig(config.Printful)
			mongo.InitPrintfulDB(config.Databases.Printful)
			mongo.InitImagesDB(config.Databases.Images)
			server.StartServer(config.HTTP)
		} else {
			log.Println("Error while reading configuration", err)
		}
	} else {
		log.Println("Error while reading configuration file", err)
	}
}
