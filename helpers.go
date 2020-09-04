package gsuite_admin_app

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

func (this App) Service(tok *oauth2.Token, ctx ...context.Context) (service *admin.Service, err error) {
	client := this.Crendentials.Client(Context(ctx...), tok)
	return admin.NewService(context.Background(), option.WithHTTPClient(client))
}

func (this App) LoadDomain(authCode string, ctx ...context.Context) (domain string, tok *oauth2.Token, err error) {
	if tok, err = this.Crendentials.Exchange(context.TODO(), authCode); err != nil {
		err = fmt.Errorf("unable to load token from auth code: %v", err)
		return
	}

	service, err := this.Service(tok, ctx...)
	if err != nil {
		err = fmt.Errorf("unable to create service %v", err)
		return
	}
	domains, err := admin.NewDomainsService(service).List(this.Customer).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve domains: %v", err)
		return
	}

	if len(domains.Domains) == 0 {
		err = errors.New("account not have domains")
		return
	}

	for _, d := range domains.Domains {
		if d.IsPrimary {
			domain = d.DomainName
			return
		}
	}

	err = fmt.Errorf("no domains found: %v", err)
	return
}

func (this App) FindDomain(tok *oauth2.Token, ctx ...context.Context) (domain string, err error) {
	service, err := this.Service(tok, ctx...)
	if err != nil {
		err = fmt.Errorf("unable to create service %v", err)
		return
	}
	domains, err := admin.NewDomainsService(service).List(this.Customer).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve domains: %v", err)
		return
	}

	if len(domains.Domains) == 0 {
		err = errors.New("account not have domains")
		return
	}

	for _, d := range domains.Domains {
		if d.IsPrimary {
			domain = d.DomainName
			return
		}
	}

	err = fmt.Errorf("no domains found: %v", err)
	return
}
