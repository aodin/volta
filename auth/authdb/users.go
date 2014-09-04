package authdb

import (
	"fmt"
	"github.com/aodin/aspect"
)

// User is a database-backed user that implements the volta auth.User
// interface.
type User struct {
	IDField       int64  `db:"id"`
	EmailField    string `db:"email"`
	PasswordField string `db:"password"`
	IsAdminField  bool   `db:"is_admin"`
	db *aspect.DB
}

// ID returns the User's id
func (u *User) ID() int64 {
	return u.IDField
}

// Email returns the User's email
func (u *User) Email() string {
	return u.EmailField
}

// Username returns the User's email
func (u *User) Username() string {
	return u.EmailField
}

func (u *User) IsAdmin() bool {
	return u.IsAdminField
}

func (u *User) Create() (err error) {
	err = u.db.Execute(Users.Insert(u))
	if err != nil {
		err = fmt.Errorf(
			"authdb: error creating user '%s': %s",
			u.EmailField,
			err,
	}
	// TODO An ID will be assigned automatically
	return
}

func (u *User) Delete() (err error) {
	err = u.db.Execute(Users.Delete(u))
	if err != nil {
		err = fmt.Errorf(
			"authdb: error deleting user '%s': %s",
			u.EmailField,
			err,
	}
	return
}

var Users = aspect.Table("users",
	aspect.Column("id", aspect.Integer{}),
	aspect.Column("email", aspect.String{}),
	aspect.Column("password", aspect.String{}),
	aspect.Column("is_admin", aspect.Boolean{}),
	aspect.PrimaryKey("id"),
)
