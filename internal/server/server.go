package server

import (
	"log"
	"os"

	"github.com/Kazalo11/six-degrees-seperation/internal/middleware"
	routes "github.com/Kazalo11/six-degrees-seperation/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func Start() {
	router := gin.Default()
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	router.Use(middleware.AuthorizeRequest())
	v1 := router.Group("/v1")
	AddRoutes(v1)
	router.Run()

}

func AddRoutes(superRoute *gin.RouterGroup) {
	routes.ArtistRoutes(superRoute)
}
