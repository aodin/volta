package auth

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/aodin/config"
	"github.com/aodin/sol"
	_ "github.com/aodin/sol/postgres" // Driver import
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testconn *sol.DB
var once sync.Once

// getConn returns a postgres connection pool
func getConn(t *testing.T) *sol.DB {
	credentials := os.Getenv("VOLTA_TEST")
	if credentials == "" {
		t.Fatalf("No testing database credentials set (VOLTA_TEST)")
	}

	once.Do(func() {
		var err error
		// TODO allow the driver to be chosen
		if testconn, err = sol.Open("postgres", credentials); err != nil {
			t.Fatalf("Failed to open connection: %s", err)
		}
		testconn.SetMaxOpenConns(20)
	})
	return testconn
}

func initSchema(conn sol.Conn, tables ...sol.Tabular) {
	// Create the given schemas
	for _, table := range tables {
		if table == nil || table.Table() == nil {
			continue
		}
		conn.Query(table.Table().Create().IfNotExists().Temporary())
	}
}

func TestAuth(t *testing.T) {
	// Create a mock auth using the testing database
	assert := assert.New(t)

	// Get a blank DB and create the schemas
	tx, _ := getConn(t).Must().Begin()
	defer tx.Rollback()
	initSchema(tx, Users, Sessions, Tokens)

	// Create a mock Auth and test its methods
	auth := Mock(config.Default, tx)

	var valid, invalid User
	var err error

	// Create a user, session, and a token
	user, err := auth.CreateUser("a@example.com", "admin", "guy", "secret")
	require.Nil(t, err, "Error during user creation")
	require.True(t, user.Exists(), "Failed to create user")

	session := auth.sessions.Create(user)
	require.True(t, session.Exists(), "Failed to create session")
	assert.Equal(user.ID, session.UserID)

	token := auth.tokens.ForeverToken(user)
	assert.Equal(user.ID, token.UserID)

	// Duplicate users cannot be created
	invalid, err = auth.CreateUser("a@example.com", "admin", "guy", "secret")
	require.NotNil(t, err, "Failed to error when creating duplicate user")
	require.False(t, invalid.Exists(), "Invalid user was created")

	// Update the user's existing token to perform auth by user token
	auth.ResetUserToken(&user)
	assert.NotEqual("", user.Token, "No user token was set")
	assert.False(user.TokenSetAt.IsZero(), "No user token timestamp was set")

	// Attempt auth by password
	valid, err = auth.ByPassword("a@example.com", "secret")
	require.Nil(t, err, "Could not auth by password")
	assert.Equal(user.ID, valid.ID)

	// Incorrect password
	_, err = auth.ByPassword("a@example.com", "1234")
	assert.NotNil(err, "Incorrect password should have errored during auth")

	// User that does not exist
	_, err = auth.ByPassword("b@example.com", "secret")
	assert.NotNil(err, "Missing email should have errored auth by password")

	// Attempt auth by session
	valid = auth.BySession(session.Key)
	assert.True(valid.Exists(), "An invalid user was returned by session key")

	// Session that does not exist
	invalid = auth.BySession("")
	assert.False(
		invalid.Exists(),
		"A valid user returned from a session that does not exist",
	)

	// Attempt auth by user token (the token field on the user schema)
	valid, err = auth.ByUserToken(user.ID, user.Token)
	assert.Nil(err, "An invalid user was returned by user token")
	assert.True(valid.Exists())

	invalid, err = auth.ByUserToken(0, user.Token)
	assert.NotNil(err, "An valid user was returned from a zero id")
	assert.False(invalid.Exists())

	invalid, err = auth.ByUserToken(user.ID, "")
	assert.NotNil(err, "An valid user was returned from an empty token")
	assert.False(invalid.Exists())

	// Attempt auth by token (used in APIs)
	valid = auth.ByToken(user.ID, token.Key)
	assert.True(valid.Exists(), "An invalid user was returned by token")

	invalid = auth.ByToken(user.ID, "")
	assert.False(invalid.Exists(), "A valid user returned from an empty token")

	// Test getter methods
	assert.NotNil(auth.Users(), "Users manager is missing")
	assert.NotNil(auth.Sessions(), "Sessions manager is missing")
	assert.NotNil(auth.Tokens(), "Tokens manager is missing")

	// Start a test server
	create := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.CreateSessionAndRedirect(w, r, user, "/redirect")
	})
	ts := httptest.NewServer(create)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	require.Nil(t, err, "Cookie jar creation errored")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("Redirects are disabled")
		},
		Jar: jar,
	}

	req, err := http.NewRequest("GET", ts.URL, nil)
	require.Nil(t, err, "Could not create a new request")

	// Send a request with a cookie jar
	// An error is expected because of the custom redirect policy
	res, err := client.Do(req)
	require.NotNil(t, err, "Test server did not redirect")
	assert.Equal(302, res.StatusCode)
	cookies := res.Cookies()
	require.Equal(t, 1, len(cookies))
	assert.Equal(auth.CookieName(), cookies[0].Name)

	location, err := res.Location()
	require.Nil(t, err, "A location header was not set")
	require.NotNil(t, location, "No location URL was returned")
	assert.Equal(ts.URL+"/redirect", location.String())

	// A session should not exist at the cookie's value
	assert.True(
		auth.Sessions().Get(cookies[0].Value).Exists(),
		"Session was not created",
	)

	// Start a test server for logout
	logout := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Error is ignored
		auth.Logout(w, r)
	})
	logoutServer := httptest.NewServer(logout)
	defer logoutServer.Close()

	// Send the same cookie jar request to logout
	res, err = client.Get(logoutServer.URL)
	require.NotNil(t, err, "Test server did not redirect")
	assert.Equal(302, res.StatusCode)

	// The session should no longer exist
	session = auth.Sessions().Get(cookies[0].Value)

	// Also attempt a logout with a cookie - it should not error
	assert.False(
		auth.Sessions().Get(cookies[0].Value).Exists(),
		"Session was not deleted",
	)
}
