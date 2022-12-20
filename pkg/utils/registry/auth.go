package registry

import (
	"net/url"

	"github.com/distribution/distribution/registry/client/auth"
)

type credstore struct {
	username      string
	password      string
	refreshTokens map[string]string
}

func (cs *credstore) Basic(*url.URL) (string, string) {
	return cs.username, cs.password
}

func (cs *credstore) RefreshToken(u *url.URL, service string) string {
	return cs.refreshTokens[service]
}

func (cs *credstore) SetRefreshToken(u *url.URL, service string, token string) {
	if cs.refreshTokens != nil {
		cs.refreshTokens[service] = token
	}
}

func NewCredStore(opts *Options) auth.CredentialStore {
	return &credstore{
		username:      opts.Username,
		password:      opts.Password,
		refreshTokens: make(map[string]string),
	}
}
