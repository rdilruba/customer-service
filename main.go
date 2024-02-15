package main

import (
	"customer-service/db"
	"customer-service/service"

	"github.com/gin-gonic/gin"
)

func main() {

	secret := db.GetSecretValue()
	db := db.GetDB(secret)

	a := service.GetApp(db)

	r := gin.Default()

	r.POST("/customers", a.PostHandler)
	r.GET("/customers/:customerId", a.GetHandler)
	r.PUT("/customers/:customerId", a.PutHandler)
	r.DELETE("/customers/:customerId", a.DeleteHandler)

	r.Run("localhost:8080")
}
