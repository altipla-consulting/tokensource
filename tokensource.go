package tokensource

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

// Notify is an oauth2.TokenSource that tracks changes in the token to be able to
// store the updated token after finishing the operations.
type Notify struct {
	*sync.Mutex
	ts      oauth2.TokenSource
	token   *oauth2.Token
	changed bool
}

// NewNotify builds a new oauth2.TokenSource that alerts when a token changes.
func NewNotify(ctx context.Context, config *oauth2.Config, token *oauth2.Token) *Notify {
	return &Notify{
		Mutex: new(sync.Mutex),
		ts:    config.TokenSource(ctx, token),
		token: token,
	}
}

// Token implements oauth2.TokenSource returning a refreshed token if needed. Any update
// will be registered to make HasChanged() return true. It is threadsafe as the
// library requires.
func (notify *Notify) Token() (*oauth2.Token, error) {
	notify.Lock()
	defer notify.Unlock()

	token, err := notify.ts.Token()
	if err != nil {
		return nil, err
	}

	if notify.token == nil || token.AccessToken != notify.token.AccessToken || token.RefreshToken != notify.token.RefreshToken {
		notify.changed = true
		notify.token = token
	}

	return token, nil
}

// HasChanged returns if we have a new token different from the one passed to NewNotify.
func (notify *Notify) HasChanged() bool {
	notify.Lock()
	defer notify.Unlock()

	return notify.changed
}

// Client builds an OAuth2-authenticated HTTP client like oauth2.Config.Client does.
func (notify *Notify) Client(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, notify)
}
