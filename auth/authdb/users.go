package authdb

import (
	"fmt"
	"github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
	"github.com/aodin/volta/auth"
)

type UserManager struct {
	db *aspect.DB
	h  auth.Hasher
}

func (m *UserManager) Create(name, password string, f ...auth.Fields) (auth.User, error) {
	var user User
	for _, fields := range f {
		if err := fields.Unmarshal(&user); err != nil {
			return &user, err
		}
		// TODO unpack the other fields? With overwrites?
		break
	}

	// Hash the password
	user.PasswordField = auth.MakePassword(m.h, password)

	// Overwrite the email with the given name (which should be email anyway)
	user.EmailField = name

	stmt := postgres.Insert(
		Users.C["email"],
		Users.C["password"],
		Users.C["is_admin"],
	).Returning(
		Users.C["id"],
	).Values(user)

	var id int64
	if err := m.db.QueryOne(stmt, &id); err != nil {
		return &user, auth.NewServerError(
			"authdb: could not create user %s: %s",
			name,
			err,
		)
	}

	// Set the new user ID
	user.IDField = id
	return &user, nil
}

func (m *UserManager) Get(fields auth.Fields) (auth.User, error) {
	// Allow queries by id or name
	rawID, ok := fields["id"]
	if ok {
		id, isInt64 := rawID.(int64)
		if !isInt64 {
			return nil, auth.NewServerError("authdb: could not parse %v as an id", rawID)
		}
		// Get the user at the given id
		var user User
		stmt := Users.Select().Where(Users.C["id"].Equals(id))
		if err := m.db.QueryOne(stmt, &user); err != nil {
			// TODO distinguish between no response and an improper query
			return &user, auth.NewUserError("no user with id %d exists", id)
		}
		return &user, nil
	}
	rawName, ok := fields["name"]
	if ok {
		name, isString := rawName.(string)
		if !isString {
			return nil, auth.NewServerError("auth: could not parse %v as a name", rawName)
		}
		// Get the user at the given name
		var user User
		stmt := Users.Select().Where(Users.C["name"].Equals(name))
		if err := m.db.QueryOne(stmt, &user); err != nil {
			// TODO distinguish between no response and an improper query
			return &user, auth.NewUserError("no user with name %s exists", name)
		}
		return &user, nil
	}
	return nil, auth.NewServerError("auth: insufficient fields provided to find a user")
}

func (m *UserManager) Delete(id int64) (auth.User, error) {
	_, err := m.db.Execute(Users.Delete().Where(Users.C["id"].Equals(id)))
	if err != nil {
		// TODO When is it a server error and when is it a user error?
		err = fmt.Errorf(
			"authdb: error deleting user with id %d: %s",
			id,
			err,
		)
	}
	return nil, nil
}

// User is a database-backed user that implements the volta auth.User
// interface.
type User struct {
	IDField       int64  `db:"id"`
	EmailField    string `db:"email"`
	PasswordField string `db:"password"`
	IsAdminField  bool   `db:"is_admin"`
}

// ID returns the User's id
func (u *User) ID() int64 {
	return u.IDField
}

// Name returns the User's email
func (u *User) Name() string {
	return u.EmailField
}

// Username returns the User's email
func (u *User) Password() string {
	return u.PasswordField
}

func (u *User) Fields() auth.Fields {
	return auth.Fields{
		"id":       u.IDField,
		"email":    u.EmailField,
		"password": u.PasswordField,
		"is_admin": u.IsAdminField,
	}
}

func (u *User) Delete() (err error) {
	return auth.NewServerError("authdb: delete users through the manager")
}

var Users = aspect.Table("users",
	aspect.Column("id", aspect.Integer{}),
	aspect.Column("email", aspect.String{}),
	aspect.Column("password", aspect.String{}),
	aspect.Column("is_admin", aspect.Boolean{}),
	aspect.PrimaryKey("id"),
)
