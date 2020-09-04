package gsuite_admin_app

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

type Token struct {
	*oauth2.Token
	Domain string
	Scopes []string `json:"scopes,omitempty"`
}

type TokenStorage interface {
	Get(r *http.Request, domain string) (token *Token, err error)
	Put(r *http.Request, token *Token) (err error)
}

type DirTokenStorage struct {
	Dir string
}

func NewDirTokenStorage(dir string) *DirTokenStorage {
	return &DirTokenStorage{Dir: dir}
}

// Retrieves a token from a local file.
func (this DirTokenStorage) Get(r *http.Request, domain string) (token *Token, err error) {
	pth := filepath.Join(this.Dir, domain, "gsuite-token.json")
	f, err := os.Open(pth)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &Token{}
	err = json.NewDecoder(f).Decode(tok)
	tok.Expiry = time.Now().Add(10 * time.Minute)
	return tok, err
}

// Retrieves a token from a local file.
func (this DirTokenStorage) Put(r *http.Request, token *Token) (err error) {
	pth := filepath.Join(this.Dir, token.Domain, "gsuite-token.json")
	f, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
