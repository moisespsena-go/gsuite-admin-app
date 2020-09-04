package gsuite_admin_app

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

func (this *App) RegisterRequestHandler(w http.ResponseWriter, r *http.Request) {
	var domain string
	switch r.Method {
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if domain = r.PostFormValue("domain"); domain == "" {
			http.Redirect(w, r, r.RequestURI, http.StatusSeeOther)
			return
		}

		this.RegisterRequest(w, r, domain)
	case http.MethodGet, http.MethodHead:
		if this.IndexHandler != nil {
			this.IndexHandler(w, r)
			return
		}
	default:
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
	}
}

func (this *App) RegisterRequest(w http.ResponseWriter, r *http.Request, domain string) {
	var (
		s RegisterSession
		err error
		scopes []string
	)
	if err := s.Load(r); err != nil {
		http.Error(w, fmt.Sprintf("unable to load register session: %v", err), http.StatusInternalServerError)
		return
	}
	if scopes, err = this.FindScopes(r); err != nil {
		http.Error(w, fmt.Sprintf("unable to find scopes: %v", err), http.StatusInternalServerError)
		return
	}

	s.Data().Scopes = scopes
	s.Data().Domain = domain

	if err := s.Save(w); err != nil {
		http.Error(w, fmt.Sprintf("unable to save register session: %v", err), http.StatusInternalServerError)
		return
	}
	if t, err := this.TokenStorage.Get(r, domain); err != nil || fmt.Sprint(t.Scopes) != fmt.Sprint(this.scopes.Values) {
		cfg := this.Crendentials
		cfg.Scopes = scopes

		if url := s.Data().StoreTokenUrl; url != "" {
			cfg.RedirectURL = url
		}
		authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		http.Redirect(w, r, authURL, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, this.DomainPageURL, http.StatusSeeOther)
	return
}

func (this *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var (
		q    = r.URL.Query()
		code = q.Get("code")
	)

	if code == "" {
		http.Error(w, "empty code param", http.StatusBadRequest)
		return
	}

	s, err := GetRegisterSession(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to load session: %v", err), http.StatusInternalServerError)
		return
	}
	if !s.HasData() {
		http.Error(w, "register session does not have data", http.StatusInternalServerError)
		return
	}

	cfg := this.Crendentials
	cfg.Scopes = s.Data().Scopes

	if url := s.Data().StoreTokenUrl; url != "" {
		cfg.RedirectURL = url
	}

	tok, err := cfg.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to load token from auth code: %v", err), http.StatusInternalServerError)
		return
	}

	var token = &Token{tok, s.Data().Domain, s.Data().Scopes}
	if err = this.TokenStorage.Put(r, token); err != nil {
		http.Error(w, "store token failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if redirectTo := s.Data().RedirectTo; redirectTo != "" {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}
}

