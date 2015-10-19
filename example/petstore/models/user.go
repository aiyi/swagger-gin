package models

import (
	"github.com/aiyi/swagger-gin/errors"
	"github.com/asaskevich/govalidator"
)

type User struct {
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"firstName,omitempty"`
	Id         int64  `json:"id,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Password   string `json:"password,omitempty"`
	Phone      string `json:"phone,omitempty"`
	UserStatus int32  `json:"userStatus,omitempty"`
	Username   string `json:"username,omitempty"`
}

func (m *User) Validate() error {
	if err := m.validateEmail(); err != nil {
		return err
	}

	return nil
}

func (m *User) validateEmail() error {
	if m.Email == "" {
		return nil
	}

	if govalidator.IsEmail(m.Email) != true {
		return errors.InvalidType("email", "body", "email", m.Email)
	}

	return nil
}
