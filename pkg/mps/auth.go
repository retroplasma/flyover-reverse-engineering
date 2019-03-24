package mps

import (
	"github.com/retroplasma/flyover-reverse-engineering/pkg/mps/auth"
)

// Session has the session ID for AuthURL
type Session struct {
	ID string
}

// TokenP1 is one of the tokens that AuthURL requires
type TokenP1 string

// AuthContext holds all additional parameters that are required for AuthURL
type AuthContext struct {
	Session
	ResourceManifest
	TokenP1
}

// AuthURL authenticates a URL
func (ctx AuthContext) AuthURL(url string) (string, error) {
	return auth.AuthURL(url, ctx.Session.ID, string(ctx.TokenP1), ctx.ResourceManifest.TokenP2)
}
