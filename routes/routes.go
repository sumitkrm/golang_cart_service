package routes

import (
	"cart/controllers"
	"fmt"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	fmt.Println("a")
	incomingRoutes.POST("/user/signup", controllers.SignUp)
	incomingRoutes.POST("/user/login", controllers.Login)
	incomingRoutes.POST("/admin/addproduct", controllers.ProductViewAdmin)
	incomingRoutes.GET("users/productview", controllers.SearchProduct)
	incomingRoutes.GET("/users/search", controllers.SearchProductByQuery)

}
