package main_test

import (
	"encoding/json"
	"go-printful-api/src/config"
	"go-printful-api/src/mongo"
	"go-printful-api/src/printful"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_, filename, _, _ := runtime.Caller(0)
	// The ".." may change depending on you folder structure
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	err = initConfig()
	if err != nil {
		panic(err)
	}
}

func RefreshAllProducts() {
	wg.Add(1)
	go func() {
		defer wg.Done()
		printful.RefreshAllProducts()
	}()
}

func initConfig() error {
	var err error
	var content []byte
	config := config.Config{}

	if content, err = os.ReadFile("config.json"); err != nil {
		return err
	}
	if err = json.Unmarshal(content, &config); err != nil {
		return err
	}
	printful.SetPrintfulConfig(config.Printful)
	mongo.InitPrintfulDB(config.Databases.Printful)
	mongo.InitImagesDB(config.Databases.Images)
	return nil
}

func TestGetProducts(t *testing.T) {
	products, err := printful.GetProducts()
	if err != nil {
		t.Error(err)
		return
	}

	j, _ := json.MarshalIndent(&products, "", "")

	err = os.WriteFile("./var/products.json", j, 0666)
	if err != nil {
		t.Error(err)
		return
	}
	//wg.Wait()
}
