package main

import (
	"github.com/aiyi/swagger-gin/example/petstore"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/v2", "C:/Users/Jesse/Work/go/src/petstore")
	r.Static("/swagger", "C:/Users/Jesse/Work/web/swagger-ui-master/dist")
	r.Static("/editor", "C:/Users/Jesse/Work/web/swagger-editor")

	petstore.Pets = r.Group("/api/pets")
	petstore.Users = r.Group("/api/users")
	petstore.Store = r.Group("/api/store")
	petstore.AddRoutes()

	r.Run(":8080")
}
