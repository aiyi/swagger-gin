package models

import "github.com/aiyi/swagger-gin/validate"

type Pet struct {
	Category  Category `json:"category,omitempty"`
	Id        int64    `json:"id,omitempty"`
	Name      string   `json:"name" binding:"required"`
	PhotoUrls []string `json:"photoUrls" binding:"required"`
	Status    string   `json:"status,omitempty"`
}

func (m *Pet) Validate() error {
	if err := m.validateName(); err != nil {
		return err
	}

	if err := m.validatePhotoUrls(); err != nil {
		return err
	}

	return nil
}

func (m *Pet) validateName() error {
	if err := validate.MaxLength("name", "body", string(m.Name), 30); err != nil {
		return err
	}

	if err := validate.MinLength("name", "body", string(m.Name), 6); err != nil {
		return err
	}

	if err := validate.Pattern("name", "body", string(m.Name), `/[a-zA-Z]/`); err != nil {
		return err
	}

	return nil
}

func (m *Pet) validatePhotoUrls() error {
	return nil
}
