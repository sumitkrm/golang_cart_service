package controllers

import (
	"cart/database"
	"cart/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	generate "cart/tokens"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type ResponseStruct struct {
	Success      bool   `json:"success"`
	ResponseCode int    `json:"response_code"`
	Message      string `json:"message"`
	//RequestId    string                 `json:"request_id"`
	//Data         map[string]interface{} `json:"data"`
}

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		fmt.Println(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, givenPassword string) (bool, string) {
	valid, msg := true, ""
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	if err != nil {
		msg = "Wrong Passsword"
		valid = false
	}
	return valid, msg
}

func SignUp(c *gin.Context) {
	fmt.Println("c")
	fmt.Println(c)
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User

	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	ValidateErr := Validate.Struct(user)
	if ValidateErr != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	fmt.Println(count)
	if err != nil {
		//log.Panic(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User email already exists"})
		return
	}

	count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	defer cancel()
	if err != nil {
		log.Panic(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is already in use"})
		return
	}

	password := HashPassword(*user.Password)
	user.Password = &password

	user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	user.User_ID = user.ID.Hex()
	token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
	user.Token = &token
	user.Refresh_Token = &refreshtoken
	user.UserCart = make([]models.ProductUser, 0)
	user.Address_Details = make([]models.Address, 0)
	user.Order_Status = make([]models.Order, 0)
	_, inserterr := UserCollection.InsertOne(ctx, user)
	if inserterr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
		return
	}
	defer cancel()
	c.JSON(http.StatusCreated, "Successfully Signed Up!!")
}

func Login(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var user models.User
	var founduser models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
		return
	}
	PasswordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
	if !PasswordIsValid {
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		fmt.Println(msg)
		return
	}
	token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
	generate.UpdateAllTokens(token, refreshToken, founduser.User_ID)
	c.JSON(http.StatusFound, founduser)

	return
}

func ProductViewAdmin(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	fmt.Println(c)
	var products models.Product
	defer cancel()
	if err := c.BindJSON(&products); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	products.Product_ID = primitive.NewObjectID()
	_, anyerr := ProductCollection.InsertOne(ctx, products)
	if anyerr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Not Created"})
		return
	}
	defer cancel()
	c.JSON(http.StatusOK, "Successfully added our Product Admin!!")

}

func SearchProduct(c *gin.Context) {
	var productlist []models.Product
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cursor, err := ProductCollection.Find(ctx, bson.D{{}})
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "Something went Wrong")
		return
	}

	err = cursor.All(ctx, &productlist)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	err = cursor.Err()
	if err != nil {
		fmt.Println("Error")
		c.IndentedJSON(400, "invalid")
		return
	}
	defer cancel()

	c.IndentedJSON(200, productlist)
	return
}

func SearchProductByQuery(c *gin.Context) {
	var SearchProducts []models.Product

	queryParam := c.Query("name")

	if queryParam == "" {
		fmt.Println("query is empty")
		c.IndentedJSON(http.StatusBadRequest, "Invalid search request")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	Searchquerydb, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParam}})
	if err != nil {
		c.IndentedJSON(404, "SOmething went wrong while fetching the data")
		return
	}

	err = Searchquerydb.All(ctx, &SearchProducts)
	if err != nil {
		fmt.Println(err)
		c.IndentedJSON(400, "searchInvalid")
	}

	defer Searchquerydb.Close(ctx)

	err = Searchquerydb.Err()
	if err != nil {
		fmt.Println(err)
		c.IndentedJSON(400, "invalid search")
		return
	}

	defer cancel()
	c.IndentedJSON(200, SearchProducts)
	return

}
