package authdb

import (
	"github.com/aodin/aspect"
	"time"
)

// Session is a database-backed session that implements the volta auth.Session
// interface.
type Session struct {
	Key     string    `db:"key"`
	UserId  int64     `db:"user_id"`
	Expires time.Time `db:"expires"`
}

var Sessions = aspect.Table("sessions",
	aspect.Column("key", aspect.String{}),
	aspect.Column("user_id", aspect.Integer{}),
	aspect.Column("expires", aspect.Timestamp{WithTimezone: true}),
	aspect.PrimaryKey("key"),
)
