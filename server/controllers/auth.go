package controllers

import (
	"net/http"
	"os"
	"time"

	"shelly/config"
	"shelly/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func handleValidationError(context *gin.Context, err error) bool {
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			switch e.Tag() {
			case "min":
				context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Password must be at least 6 characters longer"})
				return true
			case "email":
				context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid email format"})
				return true
			}
		}
	}
	return false
}

func Signup(context *gin.Context) {
	var user models.User

	if err := context.ShouldBindJSON(&user); err != nil {
		if !handleValidationError(context, err) {
			context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		}
		return
	}

	var existingUser models.User
	err := config.UserCollection.FindOne(context, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "User already exists"})
		return
	}

	if err := user.HashPassword(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Could not hash password"})
		return
	}

	result, err := config.UserCollection.InsertOne(context, user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Could not create user"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  result.InsertedID.(primitive.ObjectID).Hex(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Could not generate token"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"user": gin.H{
			"token": tokenString,
			"id":    result.InsertedID.(primitive.ObjectID).Hex(),
			"email": user.Email,
		},
	})
}

func Login(context *gin.Context) {
	var user models.User
	var foundUser models.User

	if err := context.ShouldBindJSON(&user); err != nil {
		if !handleValidationError(context, err) {
			context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		}
		return
	}

	err := config.UserCollection.FindOne(context, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "User not found"})
		return
	}

	if err := foundUser.CheckPassword(user.Password); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  foundUser.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Could not generate token"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":    foundUser.ID.Hex(),
			"token": tokenString,
			"email": foundUser.Email,
		},
	})
}
