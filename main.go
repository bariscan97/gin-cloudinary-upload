package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func UploadToCloudinary(file *multipart.FileHeader) (string, error) {

	defer func() {
		os.RemoveAll("uploads/")
	}()

	SplittedFileName := strings.Split(file.Filename, ".")
	suffix := SplittedFileName[len(SplittedFileName)-1]

	if suffix != "png" && suffix != "jpg" {
		return "", fmt.Errorf("Unexpected file")
	}
	cloudinary_url := os.Getenv("CLOUDINARY_URL")

	cld, _ := cloudinary.NewFromURL(cloudinary_url)

	ctx := context.Background()

	resp, err := cld.Upload.Upload(ctx,
		"uploads/"+file.Filename,
		uploader.UploadParams{})

	if err != nil {

		log.Fatal(err)
		return "", err
	}

	return resp.URL, nil
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.SetTrustedProxies(nil)

	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.GET("/ping", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})

	})

	router.POST("/upload", func(c *gin.Context) {

		form, err := c.MultipartForm()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		files := form.File["files"]

		images := make(map[string]string)

		for _, file := range files {

			fmt.Println(file.Filename)

			err := c.SaveUploadedFile(file, "uploads/"+file.Filename)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
				return
			}

			imgUrl, err := UploadToCloudinary(file)

			if err != nil {
				images[file.Filename] = err.Error()
			} else {
				images[file.Filename] = imgUrl
			}

		}

		c.JSON(http.StatusOK, images)
	})

	router.Run(os.Getenv("PORT"))
}
