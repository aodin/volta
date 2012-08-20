package users

type User interface {
	GetId() int
	GetUsername() string
	GetEmail() string
	Save() error
	Delete() error
}

type BaseUser struct {
	Id int
	Username string
	Email string
}