package server

import (
	"github.com/Kazalo11/six-degrees-seperation/internal/middleware"
	routes "github.com/Kazalo11/six-degrees-seperation/internal/routes"
	"github.com/gin-gonic/gin"
)

func Start() {
	router := gin.Default()
	router.Use(middleware.AuthorizeRequest())
	v1 := router.Group("/v1")
	AddRoutes(v1)
	router.Run()

}

func AddRoutes(superRoute *gin.RouterGroup) {
	routes.ArtistRoutes(superRoute)
}
