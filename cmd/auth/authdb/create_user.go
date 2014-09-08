package main

import (
	"crypto/sha1"
	"github.com/aodin/aspect"
	_ "github.com/aodin/aspect/postgres"
	"github.com/aodin/volta/auth"
	"github.com/aodin/volta/auth/authdb"
	"log"
)

func main() {
	db, err := aspect.Connect(
		"postgres",
		"host=localhost port=5432 dbname=volta user=postgres password=gotest",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hasher := auth.NewPBKDF2Hasher("test", 1, sha1.New)
	users := authdb.NewUserManager(db, hasher)

	// Create a new user
	_, err = users.Create("admin", "admin", auth.Fields{"is_admin": true})
	if err != nil {
		log.Fatalf("Unable to create admin user: %s", err)
	}
	log.Println("User created")
}
