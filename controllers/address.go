package controllers

import (
	"cart/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress(c *gin.Context) {
	user_id := c.Query("id")
	if user_id == "" {
		c.JSON(http.StatusNotFound, "Invalid Id")
	}

	address, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		c.IndentedJSON(500, "Internal Server Error")
	}

	var addresses models.Address
	addresses.Address_id = primitive.NewObjectID()
	if err = c.BindJSON(&addresses); err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, err.Error())
	}
	fmt.Println(addresses)
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	match_filter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
	group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "id", Value: "$address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}

	pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{match_filter, unwind, group})
	if err != nil {
		c.IndentedJSON(500, "Internal Server Error")
	}

	var addressinfo []bson.M
	if err = pointcursor.All(ctx, &addressinfo); err != nil {
		//panic(err)
		c.IndentedJSON(500, "Internal Server Error")
	}

	var size int32
	for _, address_no := range addressinfo {
		count := address_no["count"]
		size = count.(int32)
	}
	if size < 2 {
		filter := bson.D{primitive.E{Key: "_id", Value: address}}
		update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err := UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		c.IndentedJSON(400, "Not Allowed ")
	}
	ctx.Done()
}

func EditHomeAddress(c *gin.Context) {
	user_id := c.Query("id")
	if user_id == "" {
		c.JSON(http.StatusNotFound, "Invalid Id")
	}

	usert_id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		c.IndentedJSON(500, http.StatusInternalServerError)
		return
	}

	// Fetching adderess from rp and binding it to the address struct
	var editaddress models.Address
	if err := c.BindJSON(&editaddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}
	fmt.Println(editaddress)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
	update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house_name", Value: editaddress.House}, {Key: "address.0.street_name", Value: editaddress.Street}, {Key: "address.0.city_name", Value: editaddress.City}, {Key: "address.0.pin_code", Value: editaddress.Pincode}}}}
	_, err = UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.IndentedJSON(500, "something Went wrong")
		return
	}
	ctx.Done()
	c.IndentedJSON(200, "Updated the homeaddress")

}

func EditWorkAddress(c *gin.Context) {

	user_id := c.Query("id")
	if user_id == "" {
		c.JSON(http.StatusNotFound, "Invalid Id")
	}

	usert_id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		c.IndentedJSON(500, "something Went wrong")
		return
	}

	var editaddress models.Address
	if err := c.BindJSON(&editaddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
	update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.1.house_name", Value: editaddress.House}, {Key: "address.1.street_name", Value: editaddress.Street}, {Key: "address.1.city_name", Value: editaddress.City}, {Key: "address.1.pin_code", Value: editaddress.Pincode}}}}
	_, err = UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.IndentedJSON(500, http.StatusInternalServerError)
		return
	}
	ctx.Done()
	c.IndentedJSON(200, "Updated the homeaddress")

}

func DeleteAddress(c *gin.Context) {
	user_id := c.Query("id")

	if user_id == "" {
		fmt.Println("Invalid request")
		c.IndentedJSON(http.StatusNotFound, "Invalid Search Index")
		return
	}

	addresses := make([]models.Address, 0)
	usert_id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		c.IndentedJSON(500, http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
	update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}

	_, err = UserCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		c.IndentedJSON(500, "Something Went Wrong")
		return
	}
	defer cancel()
	ctx.Done()
	c.IndentedJSON(200, "Successfully Updated the Home address")
	return
}
