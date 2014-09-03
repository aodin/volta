package auth

type User interface {
	Id() int64
	Username() string
	Email() string
	Save() error
	Delete() error
}

// For the base user, the email is the username
type BaseUser struct {
	id       int64
	email    string
}

func (user *BaseUser) Id() int64 {
	return user.id
}

func (user *BaseUser) Username() string {
	return user.email
}

func (user *BaseUser) Email() string {
	return user.email
}

// You can't save or delete base users
func (user *BaseUser) Save() error {
	return ErrNotImplemented
}

func (user *BaseUser) Delete() error {
	return ErrNotImplemented
}

