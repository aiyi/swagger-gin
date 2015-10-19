package models

import (
	"encoding/json"
	"time"

	"github.com/aiyi/swagger-gin/errors"
	"github.com/aiyi/swagger-gin/validate"
	"github.com/asaskevich/govalidator"
)

type Order struct {
	Complete bool      `json:"complete,omitempty"`
	Contact  string    `json:"contact,omitempty"`
	Id       int64     `json:"id" binding:"required"`
	PetId    int64     `json:"petId,omitempty"`
	Quantity int32     `json:"quantity,omitempty"`
	ShipDate time.Time `json:"shipDate,omitempty"`
	Status   string    `json:"status,omitempty"`
}

func (m *Order) Validate() error {
	if err := m.validateContact(); err != nil {
		return err
	}

	if err := m.validateId(); err != nil {
		return err
	}

	if err := m.validatePetId(); err != nil {
		return err
	}

	if err := m.validateQuantity(); err != nil {
		return err
	}

	if err := m.validateStatus(); err != nil {
		return err
	}

	return nil
}

func (m *Order) validateContact() error {
	if m.Contact == "" {
		return nil
	}

	if govalidator.IsEmail(m.Contact) != true {
		return errors.InvalidType("contact", "body", "email", m.Contact)
	}

	return nil
}

func (m *Order) validateId() error {
	return nil
}

func (m *Order) validatePetId() error {
	if m.PetId == 0 {
		return nil
	}

	if err := validate.Minimum("petId", "body", float64(m.PetId), 10, false); err != nil {
		return err
	}

	return nil
}

var orderQuantityEnum []interface{}

func (m *Order) validateQuantityEnum(path, location string, value int32) error {
	if orderQuantityEnum == nil {
		var res []int32
		if err := json.Unmarshal([]byte(`[1,2,3]`), &res); err != nil {
			return err
		}
		for _, v := range res {
			orderQuantityEnum = append(orderQuantityEnum, v)
		}
	}
	if err := validate.Enum(path, location, value, orderQuantityEnum); err != nil {
		return err
	}

	return nil
}

func (m *Order) validateQuantity() error {
	if m.Quantity == 0 {
		return nil
	}

	if err := validate.MultipleOf("quantity", "body", float64(m.Quantity), 2); err != nil {
		return err
	}

	if err := validate.Minimum("quantity", "body", float64(m.Quantity), 1, true); err != nil {
		return err
	}

	if err := validate.Maximum("quantity", "body", float64(m.Quantity), 10, true); err != nil {
		return err
	}

	if err := m.validateQuantityEnum("quantity", "body", m.Quantity); err != nil {
		return err
	}

	return nil
}

var orderStatusEnum []interface{}

func (m *Order) validateStatusEnum(path, location string, value string) error {
	if orderStatusEnum == nil {
		var res []string
		if err := json.Unmarshal([]byte(`["suspend","shipment","received"]`), &res); err != nil {
			return err
		}
		for _, v := range res {
			orderStatusEnum = append(orderStatusEnum, v)
		}
	}
	if err := validate.Enum(path, location, value, orderStatusEnum); err != nil {
		return err
	}

	return nil
}

func (m *Order) validateStatus() error {
	if m.Status == "" {
		return nil
	}

	if err := m.validateStatusEnum("status", "body", m.Status); err != nil {
		return err
	}

	return nil
}
