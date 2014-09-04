package auth

// User is an interface for users.
type User interface {
	ID() int64
	Email() string
	Username() string
	IsAdmin() bool
	Create() error
	Delete() error
	// TODO Permissions
}
