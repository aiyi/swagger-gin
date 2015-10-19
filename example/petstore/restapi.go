package petstore

import (
	"net/http"
	"strconv"

	"github.com/aiyi/swagger-gin/example/petstore/models"
	"github.com/aiyi/swagger-gin/example/petstore/operations"
	"github.com/gin-gonic/gin"
)

var (
	Users *gin.RouterGroup
	Pets  *gin.RouterGroup
	Store *gin.RouterGroup
)

func AddRoutes() {
	Pets.POST("", AddPetHandler)
	Pets.PUT("", UpdatePetHandler)
	Pets.POST("/pet", UpdatePetWithFormHandler)
	Pets.GET("/pet", GetPetByIdHandler)
	Pets.DELETE("/pet", DeletePetHandler)

	Store.POST("/order", PlaceOrderHandler)
	Store.GET("/order/getOrderById", GetOrderByIdHandler)
	Store.DELETE("/order/getOrderById", DeleteOrderHandler)

	Users.POST("", CreateUserHandler)
	Users.GET("/auth/login", LoginUserHandler)
	Users.GET("/auth/logout", LogoutUserHandler)
	Users.GET("/user", GetUserByNameHandler)
	Users.PUT("/user", UpdateUserHandler)
	Users.DELETE("/user", DeleteUserHandler)

}

func AddPetHandler(c *gin.Context) {
	var body models.Pet

	if err := c.BindJSON(&body); err != nil {
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if err := operations.AddPet(&body); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func UpdatePetHandler(c *gin.Context) {
	var body models.Pet

	if err := c.BindJSON(&body); err != nil {
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if err := operations.UpdatePet(&body); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func UpdatePetWithFormHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	petId := queryValues.Get("petId")
	if petId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "petId"})
		return
	}

	name := c.Request.PostFormValue("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "name"})
		return
	}

	status := c.Request.PostFormValue("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "status"})
		return
	}

	if err := operations.UpdatePetWithForm(petId, name, status); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func GetPetByIdHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	strPetId := queryValues.Get("petId")
	if strPetId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "petId"})
		return
	}

	var petId int64
	if i, err := strconv.ParseInt(strPetId, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"invalid": "petId"})
		return
	} else {
		petId = int64(i)
	}

	if resp, err := operations.GetPetById(petId); err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func DeletePetHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	strPetId := queryValues.Get("petId")
	if strPetId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "petId"})
		return
	}

	var petId int64
	if i, err := strconv.ParseInt(strPetId, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"invalid": "petId"})
		return
	} else {
		petId = int64(i)
	}

	if err := operations.DeletePet(petId); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func GetOrderByIdHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	orderId := queryValues.Get("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "orderId"})
		return
	}

	if resp, err := operations.GetOrderById(orderId); err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func DeleteOrderHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	orderId := queryValues.Get("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "orderId"})
		return
	}

	if err := operations.DeleteOrder(orderId); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func PlaceOrderHandler(c *gin.Context) {
	var body models.Order

	if err := c.BindJSON(&body); err != nil {
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if resp, err := operations.PlaceOrder(&body); err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func LoginUserHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	username := queryValues.Get("username")

	password := queryValues.Get("password")

	if err := operations.LoginUser(username, password); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func LogoutUserHandler(c *gin.Context) {
	if err := operations.LogoutUser(); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func GetUserByNameHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	username := queryValues.Get("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "username"})
		return
	}

	if resp, err := operations.GetUserByName(username); err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func UpdateUserHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	username := queryValues.Get("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "username"})
		return
	}

	var body models.User

	if err := c.BindJSON(&body); err != nil {
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if err := operations.UpdateUser(username, &body); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func DeleteUserHandler(c *gin.Context) {
	queryValues := c.Request.URL.Query()

	username := queryValues.Get("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"missing": "username"})
		return
	}

	if err := operations.DeleteUser(username); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func CreateUserHandler(c *gin.Context) {
	var body models.User

	if err := c.BindJSON(&body); err != nil {
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if err := operations.CreateUser(&body); err == nil {
		c.String(http.StatusOK, "Success")
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}
