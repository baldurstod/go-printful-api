package main

import (
	"encoding/json"
	"go-printful-api/src/config"
	"go-printful-api/src/database"
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
			database.InitPrintfulDB(config.Databases.Printful)
			database.InitImagesDB(config.Databases.Images)
			defer database.ClosePostgre()
			server.StartServer(config.HTTP)
		} else {
			log.Println("Error while reading configuration", err)
		}
	} else {
		log.Println("Error while reading configuration file", err)
	}
}
