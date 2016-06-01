package auth

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/aodin/config"
	"github.com/aodin/sol"
)

type Auth struct {
	conn     sol.Conn
	config   config.Config
	users    *UserManager
	sessions *SessionManager
	tokens   *TokenManager
	homeURL  string

	// For testing
	now func() time.Time
}

// ByPassword attempts to authenticate the given email using the given
// cleartext password. On failure, a specific error will be returned.
func (auth *Auth) ByPassword(email, password string) (user User, err error) {
	// Get the user by email - emails MUST be unique
	if user, err = auth.users.GetByEmail(email); err != nil {
		user = User{} // Do not leak user information
		return
	}

	// Check the cleartext versus encrypted password
	if !CheckPassword(auth.users.Hasher(), password, user.Password) {
		user = User{} // Do not leak user information
		err = fmt.Errorf("auth: incorrect password for user %s", user.Email)
	}
	return
}

// BySession returns an authenticated user if the given session is valid
func (auth *Auth) BySession(key string) (user User) {
	session := auth.sessions.Get(key)
	if !session.Exists() {
		return
	}
	if !session.Expires.After(auth.now()) {
		return
	}
	user, _ = auth.users.GetByID(session.UserID)
	return
}

// ByToken returns an authenticated user if the given token is valid for the
// given user id. Tokens are used for API access.
func (auth *Auth) ByToken(id int64, key string) (user User) {
	// Do not query the database directly for the token key because that could
	// leak information through a timing attack - B-trees, yo
	for _, token := range auth.tokens.All(id) {
		if len(key) != len(token.Key) {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(key), []byte(token.Key)) != 1 {
			continue
		}
		// Expires is optional, check if it exists before checking if expired
		if token.Expires != nil && token.Expires.After(auth.now()) {
			continue
		}
		user, _ = auth.users.GetByID(token.UserID)
		break
	}
	return
}

// ByUserToken returns an authenticated user if the given user's token
// matches the given token. Companies are not added as this method
// is used only for password resets and initial account creation.
// The user token also is attached to the user's model, not the separate
// tokens table, which is used for API access.
func (auth *Auth) ByUserToken(id int64, key string) (user User, err error) {
	// TODO get the user through the user manager?
	stmt := Users.Select().Where(Users.C("id").Equals(id))
	auth.conn.Query(stmt, &user)
	if !user.Exists() {
		err = fmt.Errorf("Invalid token")
		return
	}

	if (len(key) != len(user.Token)) || subtle.ConstantTimeCompare([]byte(key), []byte(user.Token)) != 1 {
		user = User{} // Don't leak user info
		err = fmt.Errorf("Invalid token")
	}
	return
}

// CookieName returns the name of the cookie used by this auth
func (auth *Auth) CookieName() string {
	return auth.config.Cookie.Name
}

// CreateUser creates a new user.
func (auth *Auth) CreateUser(email, first, last, clear string) (User, error) {
	return auth.users.Create(email, first, last, clear)
}

// CreateSession creates a new session for the given user and redirects to
// the given next URL.
func (auth *Auth) CreateSession(w http.ResponseWriter, user User) error {
	session := auth.sessions.Create(user)
	if !session.Exists() {
		return fmt.Errorf("auth: could not create new session")
	}
	SetCookie(w, auth.config.Cookie, session)
	return nil
}

func (auth *Auth) CreateSessionAndRedirect(w http.ResponseWriter, r *http.Request, user User, next string) error {
	if err := auth.CreateSession(w, user); err != nil {
		return err
	}

	// If no next variable was provided, default to /
	if next == "" {
		next = auth.homeURL
	}
	http.Redirect(w, r, next, 302)
	return nil
}

// Logout removes the auth cookie's session key from the database
func (auth *Auth) Logout(w http.ResponseWriter, r *http.Request) error {
	// Remove the session
	cookie, err := r.Cookie(auth.CookieName())
	if err != nil {
		return nil
	}
	auth.sessions.Delete(cookie.Value)

	// TODO Remove all sessions for this user? Global Logout?
	// TODO delete the cookie?

	http.Redirect(w, r, auth.homeURL, 302)
	return nil
}

// ResetUserToken generates a new user token and resets the token timestamp.
func (auth *Auth) ResetUserToken(user *User) {
	user.Token = RandomKey()
	user.TokenSetAt = time.Now()

	// Update the user before generating an email
	stmt := Users.Update().Values(
		sol.Values{"token": user.Token, "token_set_at": user.TokenSetAt},
	).Where(Users.C("id").Equals(user.ID))
	auth.conn.Query(stmt)
}

// MakePassword returns an encrypted string of the given cleartext password
// using the auth user hasher.
func (auth *Auth) MakePassword(cleartext string) string {
	return MakePassword(auth.users.Hasher(), cleartext)
}

// Users returns the internal user manager
func (auth *Auth) Users() *UserManager {
	return auth.users
}

// Sessions returns the internal session manager
func (auth *Auth) Sessions() *SessionManager {
	return auth.sessions
}

// Tokens returns the internal token manager
func (auth *Auth) Tokens() *TokenManager {
	return auth.tokens
}

// New creates a new auth with users, sessions, and tokens
func New(c config.Config, conn sol.Conn) *Auth {
	return create(c, conn, NewUsers(conn))
}

// Mock creates a mock auth with mock users
func Mock(c config.Config, conn sol.Conn) *Auth {
	return create(c, conn, MockUsers(conn))
}

func create(c config.Config, conn sol.Conn, users *UserManager) *Auth {
	return &Auth{
		conn:     conn,
		config:   c,
		users:    users,
		sessions: NewSessions(c.Cookie, conn),
		tokens:   NewTokens(conn),
		homeURL:  "/", // TODO Set this using the given config
		now:      func() time.Time { return time.Now().In(time.UTC) },
	}
}
