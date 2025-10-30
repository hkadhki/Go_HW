package models

type Validatable interface {
	Validate() error
}
