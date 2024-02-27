package main

import (
	"cart/controllers"
	"cart/database"
	"cart/middleware"
	"cart/routes"
	"log"

	//"cart/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "User"))

	router := gin.New()
	router.Use(gin.Logger())
	//router.Use(gzip.Gzip(gzip.DefaultCompression))

	//router.POST("/user/signup", controllers.SignUp)
	//router.POST("/user/login", controllers.Login)
	routes.UserRoutes(router)
	router.Use(middleware.Authentication)
	//router.POST("/admin/addproduct", controllers.ProductViewAdmin)
	//router.GET("users/productview", controllers.SearchProduct)
	//router.GET("/users/search", controllers.SearchProductByQuery)

	router.GET("/addtocart", app.AddToCart)
	router.GET("/removeitem", app.RemoveItem)
	router.GET("/listcart", controllers.GetItemFromCart)
	router.POST("/addaddress", controllers.AddAddress)
	router.PUT("/edithomeaddress", controllers.EditHomeAddress)
	router.PUT("/editworkaddress", controllers.EditWorkAddress)
	router.GET("/deleteaddresses", controllers.DeleteAddress)
	router.GET("/cartcheckout", app.BuyFromCart)
	router.GET("/instantbuy", app.InstantBuy)
	log.Fatal(router.Run("localhost:8000"))

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	//fmt.Println("Starting on port 8090!")
	//router.Run("localhost:8090")

}
