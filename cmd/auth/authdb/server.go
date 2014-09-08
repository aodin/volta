package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/aodin/aspect"
	_ "github.com/aodin/aspect/postgres"
	"github.com/aodin/volta/auth"
	"github.com/aodin/volta/auth/authdb"
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
	db        *aspect.DB
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

	// Error is ignored because all 500 errors will be logged anyway
	if returnNow, _ := auth.Login(w, r, user, pass, srv.sessions, srv.users, srv.hasher, srv.config.Cookie, "/"); returnNow {
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

func New(db *aspect.DB, c config.Config) (*Server, error) {
	srv := &Server{
		config: c,
		db:     db,
	}
	// Create the templates
	var err error
	if srv.templates, err = ParseTemplates(c); err != nil {
		return srv, fmt.Errorf("Unable to parse templates: %s", err)
	}

	srv.hasher = auth.NewPBKDF2Hasher("test", 1, sha1.New)
	srv.users = authdb.NewUserManager(db, srv.hasher)

	// Create a new sessions manager
	srv.sessions = authdb.NewSessionManager(db, c.Cookie)

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
	db, err := aspect.Connect(
		"postgres",
		"host=localhost port=5432 dbname=volta user=postgres password=gotest",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var c = config.DefaultConfig("donotusemeasasecretkey")
	c.TemplateDir = "../templates"
	c.StaticDir = "../static"

	server, err := New(db, c)
	if err != nil {
		log.Fatal(err)
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
