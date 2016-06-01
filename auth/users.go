package auth

import (
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/aodin/sol"
	"github.com/aodin/sol/postgres"
	"github.com/aodin/sol/types"
)

// User is a database-backed user.
type User struct {
	ID          int64        `db:"id,omitempty"`
	Email       string       `db:"email"`
	FirstName   string       `db:"first_name"`
	LastName    string       `db:"last_name"`
	About       string       `db:"about"`
	Photo       string       `db:"photo"`
	IsActive    bool         `db:"is_active"`
	IsSuperuser bool         `db:"is_superuser"`
	Password    string       `db:"password"`
	Token       string       `db:"token"`
	TokenSetAt  time.Time    `db:"token_set_at,omitempty"`
	CreatedAt   time.Time    `db:"created_at,omitempty"`
	manager     *UserManager `db:"-"`
}

// Delete removes the user with the given ID from the database.
// It will return an error if the user does not have an ID or the ID
// was not deleted from the database. It will panic on any connection error.
func (user User) Delete() error {
	if !user.Exists() {
		return fmt.Errorf("auth: users without IDs cannot be deleted")
	}
	return user.manager.Delete(user.ID)
}

// Exists returns true if the user has an assigned ID
func (user User) Exists() bool {
	return user.ID != 0
}

// Name returns the concatenated first and last name
func (user User) Name() string {
	return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
}

// String returns the user id and email
func (user User) String() string {
	return fmt.Sprintf("%d: %s", user.ID, user.Email)
}

// Users is the postgres schema for users
var Users = postgres.Table("users",
	sol.Column("id", postgres.Serial()),
	sol.Column("email", types.Varchar().Limit(256).NotNull()),
	sol.Column("first_name", types.Varchar().Limit(64).NotNull()),
	sol.Column("last_name", types.Varchar().Limit(64).NotNull()),
	sol.Column("about", types.Varchar().Limit(512).NotNull()),
	sol.Column("photo", types.Varchar().Limit(512).NotNull()),
	sol.Column("is_active", types.Boolean().NotNull().Default(true)),
	sol.Column("is_superuser", types.Boolean().NotNull().Default(false)),
	sol.Column("password", types.Varchar().Limit(256).NotNull()),
	sol.Column("token", types.Varchar().Limit(256).NotNull()),
	sol.Column(
		"token_set_at",
		postgres.Timestamp().WithTimezone().NotNull().Default(postgres.Now),
	),
	sol.Column(
		"created_at",
		postgres.Timestamp().WithTimezone().NotNull().Default(postgres.Now),
	),
	sol.PrimaryKey("id"),
	sol.Unique("email"),
)

// UserManager is the internal manager of users
type UserManager struct {
	conn      sol.Conn
	hash      Hasher
	tokenFunc KeyFunc
}

// Create will create a new user with given email and cleartext password.
// It will panic on any crypto or database connection errors.
func (m *UserManager) Create(email, first, last, clear string) (User, error) {
	return m.create(email, first, last, clear, false)
}

// CreateSuperuser will create a new superuser with given email and cleartext password.
// It will panic on any crypto or database connection errors.
func (m *UserManager) CreateSuperuser(email, first, last, clear string) (User, error) {
	return m.create(email, first, last, clear, true)
}

func (m *UserManager) create(email, first, last, clear string, isAdmin bool) (User, error) {
	user := User{
		Email:       email,
		FirstName:   first,
		LastName:    last,
		IsActive:    true,
		IsSuperuser: isAdmin,
		Password:    MakePassword(m.hash, clear),
		Token:       m.tokenFunc(),
		TokenSetAt:  time.Now(),
		manager:     m,
	}
	err := m.createUser(&user)
	return user, err
}

// createUser checks for a duplicate email before inserting the user.
// Email must already be normalized.
func (m *UserManager) createUser(user *User) error {
	email := sol.Select(
		Users.C("email"),
	).Where(Users.C("email").Equals(user.Email)).Limit(1)

	var duplicate string
	m.conn.Query(email, &duplicate)
	if duplicate != "" {
		return fmt.Errorf(
			"auth: user with email %s already exists", duplicate,
		)
	}

	// Insert the new user
	m.conn.Query(postgres.Insert(Users).Values(user).Returning(), user)
	return nil
}

// Delete removes the user with the given ID from the database.
// It will return an error if the ID was not deleted from the database.
// It will panic on any connection error.
func (m *UserManager) Delete(id int64) error {
	stmt := Users.Delete().Where(Users.C("id").Equals(id))
	return m.conn.Query(stmt)
}

// GetByEmail returns the user with the given email.
func (m *UserManager) GetByEmail(email string) (user User, err error) {
	stmt := Users.Select().Where(Users.C("email").Equals(email))
	if err = m.conn.Query(stmt, &user); err != nil {
		return
	}
	if !user.Exists() {
		err = fmt.Errorf("auth: no user with email %s exists", email)
	}
	return
}

// GetByID returns the user with the given id.
func (m *UserManager) GetByID(id int64) (user User, err error) {
	stmt := Users.Select().Where(Users.C("id").Equals(id))
	if err = m.conn.Query(stmt, &user); err != nil {
		return
	}
	if !user.Exists() {
		err = fmt.Errorf("auth: no user with id %d exists", id)
	}
	return
}

// Hasher returns the hasher used by the UserManager
func (m UserManager) Hasher() Hasher {
	return m.hash
}

func NewUsers(conn sol.Conn) *UserManager {
	hasher, err := GetHasher("pbkdf2_sha256")
	if err != nil {
		log.Panicf("auth: could not get pbkdf2_sha256 hasher: %s", err)
	}
	return newUsers(conn, hasher)
}

func MockUsers(conn sol.Conn) *UserManager {
	return newUsers(conn, MockHasher("mock", 1, sha1.New))
}

func newUsers(conn sol.Conn, hash Hasher) *UserManager {
	return &UserManager{
		conn:      conn,
		hash:      hash,
		tokenFunc: RandomKey,
	}
}
