package authmem

import (
	"fmt"
	"sync"
)

// UserManager is an in-memory store of users
type UserManager struct {
	mutex sync.RWMutex
	byID  map[int64]*User
	n     int64
}

func (m *UserManager) NewUser(name string, isAdmin bool) (*User, error) {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create a new user and assign an id
	m.n += 1
	u := &User{id: m.n, email: name, isAdmin: isAdmin, manager: m}
	if _, exists := m.byID[u.id]; exists {
		return u, fmt.Errorf("authmem: user with id %d already exists", u.id)
	}
	m.byID[u.id] = u
	return u, nil
}

func (m *UserManager) GetUser(id int64) (*User, error) {
	// Lock the mutex for reading
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get the user at the given id
	user, exists := m.byID[id]
	if !exists {
		return user, fmt.Errorf("authmem: no user with id %d exists", id)
	}
	return user, nil
}

func (m *UserManager) DeleteUser(id int64) error {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.byID, id)
	return nil
}

// NewUserManager initializes a new UserManager
func NewUserManager() *UserManager {
	return &UserManager{
		byID: make(map[int64]*User),
	}
}

// User is an in-memory implementation of the volta.auth User interface.
type User struct {
	id      int64
	email   string
	isAdmin bool
	manager *UserManager
}

// ID returns the User's id
func (u *User) ID() int64 {
	return u.id
}

// Email returns the User's email
func (u *User) Email() string {
	return u.email
}

// Username returns the User's email
func (u *User) Username() string {
	return u.email
}

// IsAdmin indicates if a User is an Admin
func (u *User) IsAdmin() bool {
	return u.isAdmin
}

// Create saves the user to its user manager. If the user's id field was set,
// it will be ignored.
// TODO this method is only useful for testing / duplicating users
func (u *User) Create() error {
	user, err := u.manager.NewUser(u.email, u.isAdmin)
	if err != nil {
		return err
	}
	*u = *user
	return nil
}

// Delete removes the user from its user manager.
func (u *User) Delete() error {
	return u.manager.DeleteUser(u.id)
}
