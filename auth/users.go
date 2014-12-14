package auth

import (
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	sql "github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
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
var Users = sql.Table("users",
	sql.Column("id", postgres.Serial{NotNull: true}),
	sql.Column("email", sql.String{Unique: true, NotNull: true, Length: 255}),
	sql.Column("first_name", sql.String{Length: 255, NotNull: true}),
	sql.Column("last_name", sql.String{Length: 255, NotNull: true}),
	sql.Column("about", sql.String{Length: 511, NotNull: true}),
	sql.Column("photo", sql.String{Length: 511, NotNull: true}),
	sql.Column("is_active", sql.Boolean{Default: sql.True}),
	sql.Column("is_superuser", sql.Boolean{Default: sql.False}),
	sql.Column("password", sql.String{Length: 255, NotNull: true}),
	sql.Column("token", sql.String{Length: 255, NotNull: true}),
	sql.Column("token_set_at", sql.Timestamp{Default: postgres.Now}),
	sql.Column("created_at", sql.Timestamp{Default: postgres.Now}),
	sql.PrimaryKey("id"),
)

// UserManager is the internal manager of users
type UserManager struct {
	conn      sql.Connection
	hash      Hasher
	tokenFunc KeyFunc
}

// Create will create a new user with given email and cleartext password.
// It will panic on any crypto or database connection errors.
func (m *UserManager) Create(email, first, last, clear string) (User, error) {
	user := User{
		Email:      email,
		FirstName:  first,
		LastName:   last,
		IsActive:   true,
		Password:   MakePassword(m.hash, clear),
		Token:      m.tokenFunc(),
		TokenSetAt: time.Now(),
		manager:    m,
	}
	err := m.create(&user)
	return user, err
}

// create checks for a duplicate email before inserting the user.
// Email must already be normalized.
func (m *UserManager) create(user *User) error {
	email := Users.Select().Where(Users.C["email"].Equals(user.Email)).Limit(1)
	if m.conn.MustQueryOne(email, user) {
		return fmt.Errorf(
			"auth: user with email %s already exists", user.Email,
		)
	}

	// Insert the new user
	stmt := postgres.Insert(Users).Returning(Users.Columns()...).Values(user)
	m.conn.MustQueryOne(stmt, user)
	return nil
}

// Delete removes the user with the given ID from the database.
// It will return an error if the ID was not deleted from the database.
// It will panic on any connection error.
func (m *UserManager) Delete(id int64) error {
	stmt := Users.Delete().Where(Users.C["id"].Equals(id))
	rowsAffected, err := m.conn.MustExecute(stmt).RowsAffected()
	if err != nil {
		return fmt.Errorf("auth: error during rows affected: %s", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("auth: user ID %d was not deleted", id)
	}
	return nil
}

// GetByEmail returns the user with the given email.
func (m *UserManager) GetByEmail(email string) (user User, err error) {
	stmt := Users.Select().Where(Users.C["email"].Equals(email))
	if !m.conn.MustQueryOne(stmt, &user) {
		err = fmt.Errorf("auth: no user with email %s exists", email)
	}
	return
}

// GetByID returns the user with the given id.
func (m *UserManager) GetByID(id int64) (user User, err error) {
	stmt := Users.Select().Where(Users.C["id"].Equals(id))
	if !m.conn.MustQueryOne(stmt, &user) {
		err = fmt.Errorf("auth: no user with id %d exists", id)
	}
	return
}

// Hasher returns the hasher used by the UserManager
func (m UserManager) Hasher() Hasher {
	return m.hash
}

func NewUsers(conn sql.Connection) *UserManager {
	hasher, err := GetHasher("pbkdf2_sha256")
	if err != nil {
		log.Panicf("auth: could not get pbkdf2_sha256 hasher: %s", err)
	}
	return newUsers(conn, hasher)
}

func MockUsers(conn sql.Connection) *UserManager {
	return newUsers(conn, MockHasher("mock", 1, sha1.New))
}

func newUsers(conn sql.Connection, hash Hasher) *UserManager {
	return &UserManager{
		conn:      conn,
		hash:      hash,
		tokenFunc: RandomKey,
	}
}
