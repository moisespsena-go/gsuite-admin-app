package gsuite_admin_app

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/gorilla/sessions"
)

var (
	cookieKey  = fmt.Sprintf("%x", sha1.Sum([]byte(pkg)))
	hashKey, _ = uuid.NewSHA1(uuid.NameSpaceX500, []byte(pkg)).MarshalBinary()
	store      = sessions.NewCookieStore([]byte("asdasdasd"))
)

func GetRegisterSession(r *http.Request) (session *RegisterSession, err error) {
	var s *sessions.Session
	if s, err = store.Get(r, "gsuite-admin-app"); err != nil {
		return
	}
	session = &RegisterSession{Session: s, Request: r}
	if s.Options == nil {
		s.Options = &sessions.Options{}
	}
	if s.Options.MaxAge == 0 {
		s.Options.MaxAge = 60
	}

	return
}

type RegisterSessionData struct {
	Domain        string   `json:"domains,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
	RedirectTo    string   `json:"redirect_to,omitempty"`
	StoreTokenUrl string   `json:"store_token_url,omitempty"`
}

type RegisterSession struct {
	Session *sessions.Session
	Request *http.Request
	data    *RegisterSessionData
}

func (this *RegisterSession) HasData() bool {
	return this.data != nil || this.Session.Values["@data"] != nil
}

func (this *RegisterSession) Get(key string, result interface{}) error {
	if v, ok := this.Session.Values[key]; ok {
		return json.Unmarshal([]byte(v.(string)), result)
	}
	return nil
}

func (this *RegisterSession) GetS(key string) string {
	return this.Session.Values[key].(string)
}

func (this *RegisterSession) Set(w http.ResponseWriter, key string, value interface{}) error {
	if _, ok := value.(string); !ok {
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		value = string(b)
	}
	this.Session.Values[key] = value

	return this.Session.Save(this.Request, w)
}

func (this *RegisterSession) Pop(w http.ResponseWriter, key string) error {
	if _, ok := this.Session.Values[key]; ok {
		delete(this.Session.Values, key)
		return this.Session.Save(this.Request, w)
	}
	return nil
}

func (this *RegisterSession) PopLoad(w http.ResponseWriter, key string, result interface{}) error {
	if _, ok := this.Session.Values[key]; ok {
		if err := this.Get(key, result); err != nil {
			return err
		}
		return this.Pop(w, key)
	}
	return nil
}

func (this *RegisterSession) Data() *RegisterSessionData {
	if this.data == nil {
		this.data = &RegisterSessionData{}
		if err := this.Get("@data", this.data); err != nil {
			panic(err)
		}
	}
	return this.data
}

func (this *RegisterSession) SetData(data *RegisterSessionData) {
	this.data = data
}

func (this *RegisterSession) Save(w http.ResponseWriter) (err error) {
	return this.Set(w, "@data", this.data)
}

func (this *RegisterSession) Delete(w http.ResponseWriter) (err error) {
	s2 := sessions.NewSession(this.Session.Store(), this.Session.Name())
	s2.Options = this.Session.Options
	return s2.Save(this.Request, w)
}

func (this *RegisterSession) Load(r *http.Request) (err error) {
	var s *RegisterSession
	if s, err = GetRegisterSession(r); err != nil {
		return
	}
	*this = *s
	return
}
