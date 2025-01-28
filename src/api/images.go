package api

import (
	"errors"
	"go-printful-api/src/mongo"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ImageHandler(c *gin.Context) {
	log.Println(c.FullPath(), c.Param("id"))

	img, err := mongo.GetImage(c.Param("id"))
	if err != nil {
		jsonError(c, errors.New("failed to read image"))
		return
	}

	c.Data(http.StatusOK, "image/png", img)
}
