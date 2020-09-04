package gsuite_admin_app

import (
	"fmt"
	"google.golang.org/api/admin/directory/v1"
	"net/http"
	"sync"

	"github.com/moisespsena-go/path-helpers"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/go-chi/chi"
)

var (
	pkg = path_helpers.GetCalledDir()
)

type ScopeAppender =func(app *App, scopes *Scopes, r *http.Request) (err error)
type SetupHandler = func(app *App, token *Token, r *http.Request) (err error)

type App struct {
	Customer      string
	TokenStorage  TokenStorage
	DomainPageURL string
	SetupURL      string
	IndexHandler  func(w http.ResponseWriter, r *http.Request)

	crendentialsMu            sync.Mutex
	Crendentials              oauth2.Config
	setupHandlers             []SetupHandler
	scopeAppenders            []ScopeAppender
	scopes                    *Scopes
	RequireDomainToStoreToken bool
}

func New(customer string) *App {
	return &App{Customer: customer, scopes: &Scopes{}}
}

func (this *App) AddScope(s ...string) {
	this.scopes.Add(s...)
	this.Crendentials.Scopes = this.scopes.Values
}

func (this *App) ScopeAppender(f ...ScopeAppender) []ScopeAppender {
	this.scopeAppenders = append(this.scopeAppenders, f...)
	return this.scopeAppenders
}

func (this *App) FindScopes(r *http.Request) (scopes[]string, err error) {
	var s = &Scopes{this.scopes.Values}
	s.Add(admin.AdminDirectoryDomainScope)
	for _, f := range this.scopeAppenders {
		if err = f(this, s, r); err != nil {
			return
		}
	}
	return s.Values, nil
}

func (this *App) SetupHandler(f ...SetupHandler) {
	this.setupHandlers = append(this.setupHandlers, f...)
}

func (this *App) Setup(token *Token, r *http.Request) (err error) {
	for _, f := range this.setupHandlers {
		if err = f(this, token, r); err != nil {
			return
		}
	}
	return
}

func (this *App) Scopes() []string {
	return this.scopes.Values
}

func (this *App) LoadCredentials(b []byte) (err error) {
	this.crendentialsMu.Lock()
	defer this.crendentialsMu.Unlock()

	var config *oauth2.Config
	// If modifying these scopes, delete your previously saved token.json.
	config, err = google.ConfigFromJSON(b, this.scopes.Values...)
	if err != nil {
		err = fmt.Errorf("Unable to parse client secret file to config: %v", err)
		return
	}

	this.Crendentials = *config
	return
}

func (this *App) Handler() http.Handler {
	mux := chi.NewMux()
	mux.HandleFunc("/", this.RegisterRequestHandler)
	mux.Get("/register", this.RegisterHandler)
	return mux
}
