// Package oauth abstracts external identity providers (Google today; Robinhood,
// GitHub, etc. plug in by implementing Provider and registering). The web layer
// depends only on this interface and the normalized Profile, so adding a provider
// never touches the auth flow.
package oauth

import "context"

// Profile is the normalized identity returned by any provider.
type Profile struct {
	ProviderUserID string // stable unique id at the provider (e.g. Google "sub")
	Email          string
	Name           string
	AvatarURL      string
}

// Provider is one OAuth2 identity source. AuthCodeURL builds the consent-screen
// URL; Exchange swaps the returned code for the user's normalized profile.
type Provider interface {
	Name() string
	AuthCodeURL(state, redirectURI string) string
	Exchange(ctx context.Context, code, redirectURI string) (Profile, error)
}

// Registry holds the configured providers by name.
type Registry map[string]Provider

func (r Registry) Get(name string) (Provider, bool) {
	p, ok := r[name]
	return p, ok
}
