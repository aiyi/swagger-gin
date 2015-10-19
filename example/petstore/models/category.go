package models

type Category struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (m *Category) Validate() error {
	return nil
}
