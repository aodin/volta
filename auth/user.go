package auth

import (
	"fmt"
	"sync"
)

// UserManager is an interface for managing users.
type UserManager interface {
	Create(name, password string, f ...Fields) (User, error)
	Get(fields Fields) (User, error) // Must return one and only one user
	Delete(id int64) (User, error)
}

// User is an interface for users.
type User interface {
	ID() int64
	Name() string
	Password() string // Should return the password hash, not the plaintext
	Fields() Fields
	Delete() error
	// TODO Permissions
}

// MemoryUsers is an in-memory store of users
type MemoryUsers struct {
	mutex  sync.RWMutex
	byName map[string]*user
	byID   map[int64]*user
	n      int64
	hasher Hasher
}

func (m *MemoryUsers) Create(name, password string, f ...Fields) (User, error) {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create a new user and assign an id
	m.n += 1
	u := &user{id: m.n, email: name, manager: m}

	// The id and name must be unique
	if _, exists := m.byID[u.id]; exists {
		return u, fmt.Errorf("auth: user with id %d already exists", u.id)
	}
	if _, exists := m.byName[u.email]; exists {
		return u, fmt.Errorf("auth: user with name %s already exists", u.email)
	}

	// TODO Extract the fields the manager cares about

	// Hash the password
	u.password = MakePassword(m.hasher, password)

	// Save the user to the id and name maps
	m.byID[u.id] = u
	m.byName[u.email] = u
	return u, nil
}

// Get returns a user either by ID or Name, with preference given to ID
func (m *MemoryUsers) Get(f Fields) (User, error) {
	// Lock the mutex for reading
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rawID, ok := f["ID"]
	if ok {
		id, isInt64 := rawID.(int64)
		if !isInt64 {
			return nil, fmt.Errorf("auth: could not parse %v as an id", rawID)
		}
		// Get the user at the given id
		user, exists := m.byID[id]
		if !exists {
			return user, fmt.Errorf("auth: no user with id %d exists", id)
		}
		return user, nil
	}
	rawName, ok := f["Name"]
	if ok {
		name, isString := rawName.(string)
		if !isString {
			return nil, fmt.Errorf("auth: could not parse %v as a name", rawName)
		}
		// Get the user at the given name
		user, exists := m.byName[name]
		if !exists {
			return user, fmt.Errorf("auth: no user with name %s exists", name)
		}
		return user, nil
	}
	return nil, fmt.Errorf("auth: insufficient fields provided to find a user")
}

func (m *MemoryUsers) Delete(id int64) (User, error) {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Confirm the user exists
	user, exists := m.byID[id]
	if !exists {
		return nil, fmt.Errorf("auth: no user found with id %d", id)
	}
	// Remove the user from both the id and name maps
	delete(m.byID, id)
	delete(m.byName, user.email)
	return user, nil
}

// UsersInMemory initializes a new MemoryUsers
func UsersInMemory(hasher Hasher) *MemoryUsers {
	return &MemoryUsers{
		byID:   make(map[int64]*user),
		byName: make(map[string]*user),
		hasher: hasher,
	}
}

// User is an in-memory implementation of the volta.auth User interface.
type user struct {
	id       int64
	email    string
	password string
	isAdmin  bool
	manager  *MemoryUsers
}

// ID returns the User's id
func (u *user) ID() int64 {
	return u.id
}

func (u *user) Name() string {
	return u.email
}

// Password returns the User's hashed password
func (u *user) Password() string {
	return u.password
}

func (u *user) Fields() Fields {
	return Fields{"Email": u.email, "IsAdmin": u.isAdmin}
}

// Delete removes the user from its user manager.
func (u *user) Delete() error {
	_, err := u.manager.Delete(u.id)
	return err
}
