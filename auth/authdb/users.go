package authdb

import (
	"github.com/aodin/aspect"
)

// User is a database-backed user that implements the volta auth.User
// interface.
type User struct {
	ID       int64  `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	IsAdmin  bool   `db:"is_admin"`
}

var Users = aspect.Table("users",
	aspect.Column("id", aspect.Integer{}),
	aspect.Column("email", aspect.String{}),
	aspect.Column("password", aspect.String{}),
	aspect.Column("is_admin", aspect.Boolean{}),
	aspect.PrimaryKey("id"),
)
