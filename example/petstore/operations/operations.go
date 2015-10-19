package operations

import "github.com/aiyi/swagger-gin/example/petstore/models"

func CreateUser(user *models.User) error {
	return nil
}

func LoginUser(username string, password string) error {
	return nil
}

func LogoutUser() error {
	return nil
}

func GetUserByName(username string) (*models.User, error) {
	return &models.User{}, nil
}

func UpdateUser(username string, user *models.User) error {
	return nil
}

func DeleteUser(username string) error {
	return nil
}

func AddPet(pet *models.Pet) error {
	return nil
}

func UpdatePet(pet *models.Pet) error {
	return nil
}

func UpdatePetWithForm(petId string, name string, status string) error {
	return nil
}

func GetPetById(petId int64) (*models.Pet, error) {
	return &models.Pet{}, nil
}

func DeletePet(petId int64) error {
	return nil
}

func PlaceOrder(order *models.Order) (*models.Order, error) {
	return &models.Order{}, nil
}

func GetOrderById(orderId string) (*models.Order, error) {
	return &models.Order{}, nil
}

func DeleteOrder(orderId string) error {
	return nil
}
