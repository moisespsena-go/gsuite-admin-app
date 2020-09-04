package gsuite_admin_app

import "sort"

type Scopes struct {
	Values []string
}

func (this *Scopes) Add(scope ...string) Scopes {
main:
	for _, scope := range scope {
		for _, el := range this.Values {
			if scope == el {
				continue main
			}
		}
		this.Values = append(this.Values, scope)
	}
	sort.Strings(this.Values)
	return *this
}
