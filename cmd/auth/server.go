package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/aodin/volta/auth"
	"github.com/aodin/volta/config"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

type Templates map[string]*template.Template

func (t Templates) Execute(name string, w io.Writer, data interface{}) error {
	tmpl, ok := t[name]
	if !ok {
		return fmt.Errorf("no template with the name %s exists", name)
	}
	return tmpl.Execute(w, data)
}

func ParseTemplates(conf config.Config) (t Templates, err error) {
	// TODO parent templates
	t = Templates{}
	// TODO Automatically load child templates according to some naming scheme
	templates := []string{
		"root.html",
		"login.html",
	}
	for _, file := range templates {
		path := filepath.Join(conf.TemplateDir, file)
		t[file], err = template.ParseFiles(path)
		if err != nil {
			return
		}
	}
	return
}

// Server is an example server implementing the in-memory sessions and users.
type Server struct {
	config    config.Config
	users     auth.UserManager
	sessions  auth.SessionManager
	hasher    auth.Hasher
	router    *httprouter.Router
	templates Templates
}

func (srv *Server) Root(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Check if the session is valid
	cookie, err := r.Cookie(srv.config.Cookie.Name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	user := auth.GetUserIfValidSession(srv.sessions, srv.users, cookie.Value)
	data := map[string]interface{}{
		"StaticURL": "/static/",
		"User":      user,
	}
	if err := srv.templates.Execute("root.html", w, data); err != nil {
		log.Printf("Error while executing Root: %s\n", err)
	}
}

// If the request is a GET display the login form
func (srv *Server) LoginForm(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// TODO Redirect if the user is already authenticated?
	data := map[string]interface{}{
		"StaticURL": "/static/",
	}
	if err := srv.templates.Execute("login.html", w, data); err != nil {
		log.Printf("Error while executing LoginForm: %s\n", err)
	}
}

func (srv *Server) Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user := r.FormValue("username")
	pass := r.FormValue("password")

	// Use an authentication block
	if func(username, password string) (ok bool) {
		// Get the user
		user, err := srv.users.Get(auth.Fields{"Name": username})
		// TODO What if it is a db error?
		if err != nil {
			return
		}

		// Test the user's password
		if !auth.CheckPassword(srv.hasher, password, user.Password()) {
			return
		}

		// Create a new session
		session, err := srv.sessions.Create(user)
		if err != nil {
			// TODO panic
			return
		}
		auth.SetCookie(w, srv.config.Cookie, session)
		ok = true
		return
	}(user, pass) {
		// Redirect
		http.Redirect(w, r, "/", 302)
		return
	}

	data := map[string]interface{}{
		"StaticURL": "/static/",
		"Message":   "Invalid credentials",
		"Username":  user,
	}
	if err := srv.templates.Execute("login.html", w, data); err != nil {
		log.Printf("Error while executing LoginForm: %s\n", err)
	}
}

func (srv *Server) Logout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Remove the session
	cookie, err := r.Cookie(srv.config.Cookie.Name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	srv.sessions.Delete(cookie.Value)

	http.Redirect(w, r, "/", 302)
	return
}

func (srv *Server) ListenAndServe() error {
	log.Printf("Starting server on %s\n", srv.config.Address())
	return http.ListenAndServe(srv.config.Address(), srv.router)
}

func New(c config.Config) (*Server, error) {
	srv := &Server{
		config: c,
	}
	// Create the templates
	var err error
	if srv.templates, err = ParseTemplates(c); err != nil {
		return srv, fmt.Errorf("Unable to parse templates: %s", err)
	}

	srv.hasher = auth.NewPBKDF2Hasher("test", 1, sha1.New)
	srv.users = auth.UsersInMemory(srv.hasher)

	// Create a new user
	_, err = srv.users.Create("admin", "admin")
	if err != nil {
		return srv, fmt.Errorf("Unable to create admin user: %s", err)
	}

	// Create a new sessions manager
	srv.sessions = auth.SessionsInMemory(c.Cookie, srv.users)

	// Create the routes
	srv.router = httprouter.New()
	srv.router.GET("/", srv.Root)
	srv.router.GET("/login", srv.LoginForm)
	srv.router.POST("/login", srv.Login)
	srv.router.GET("/logout", srv.Logout)
	srv.router.ServeFiles("/static/*filepath", http.Dir(c.StaticDir))

	return srv, nil
}

func main() {
	var c = config.DefaultConfig("donotusemeasasecretkey")
	c.TemplateDir = "./templates"
	c.StaticDir = "./static"

	server, err := New(c)
	if err != nil {
		log.Fatal(err)
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
