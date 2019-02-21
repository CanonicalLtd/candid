// Copyright 2018 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package static contains identity providers that validate against a static list of users.
// This provider is only intended for testing purposes.
package static

import (
	"context"
	"net/http"
	"strings"

	"gopkg.in/CanonicalLtd/candidclient.v1/params"
	"gopkg.in/errgo.v1"
	"gopkg.in/macaroon-bakery.v2/httpbakery"

	"github.com/CanonicalLtd/candid/idp"
	"github.com/CanonicalLtd/candid/idp/idputil"
	"github.com/CanonicalLtd/candid/store"
)

func init() {
	idp.Register("static", func(unmarshal func(interface{}) error) (idp.IdentityProvider, error) {
		var p Params
		if err := unmarshal(&p); err != nil {
			return nil, errgo.Notef(err, "cannot unmarshal static parameters")
		}
		if p.Name == "" {
			p.Name = "static"
		}

		return NewIdentityProvider(p), nil
	})
}

type Params struct {
	// Name is the name that will be given to the identity provider.
	Name string `yaml:"name"`

	// Domain is the domain with which all identities created by this
	// identity provider will be tagged (not including the @ separator).
	Domain string `yaml:"domain"`

	// Users is the set of users that are allowed to authenticate, with their
	// passwords and list of groups.
	Users map[string]UserInfo `yaml:"users"`
}

type UserInfo struct {
	// Password is the password for the user.
	Password string `yaml:"password"`
	// Name is the full name of the user.
	Name string `yaml:"name"`
	// Email is the user e-mail.
	Email string `yaml:"email"`
	// Groups is the list of groups the user belongs to.
	Groups []string `yaml:"groups"`
}

// NewIdentityProvider creates a new static identity provider.
func NewIdentityProvider(p Params) idp.IdentityProvider {
	return &identityProvider{params: p}

}

type identityProvider struct {
	params     Params
	initParams idp.InitParams
}

// Name implements idp.IdentityProvider.Name.
func (idp *identityProvider) Name() string {
	return idp.params.Name
}

// Domain implements idp.IdentityProvider.Domain.
func (idp *identityProvider) Domain() string {
	return idp.params.Domain
}

// Description implements idp.IdentityProvider.Description.
func (idp *identityProvider) Description() string {
	return "Static identity provider"
}

// Interactive implements idp.IdentityProvider.Interactive.
func (*identityProvider) Interactive() bool {
	return true
}

// Init implements idp.IdentityProvider.Init.
func (idp *identityProvider) Init(ctx context.Context, params idp.InitParams) error {
	idp.initParams = params
	return nil
}

// URL implements idp.IdentityProvider.URL.
func (idp *identityProvider) URL(dischargeID string) string {
	return idputil.URL(idp.initParams.URLPrefix, "/login", dischargeID)
}

// SetInteraction implements idp.IdentityProvider.SetInteraction.
func (idp *identityProvider) SetInteraction(ierr *httpbakery.Error, dischargeID string) {
}

//  GetGroups implements idp.IdentityProvider.GetGroups.
func (idp *identityProvider) GetGroups(ctx context.Context, identity *store.Identity) ([]string, error) {
	_, fulluser := identity.ProviderID.Split()
	username := strings.SplitN(fulluser, "@", 2)[0]
	if user, ok := idp.params.Users[username]; ok {
		groups := make([]string, len(user.Groups))
		for i, group := range user.Groups {
			groups[i] = idputil.NameWithDomain(group, idp.params.Domain)
		}
		return groups, nil
	}
	return []string{}, nil
}

// Handle implements idp.IdentityProvider.Handle.
func (idp *identityProvider) Handle(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	switch strings.TrimPrefix(req.URL.Path, idp.initParams.URLPrefix) {
	case "/login":
		if err := idp.handleLogin(ctx, w, req); err != nil {
			idp.initParams.VisitCompleter.Failure(ctx, w, req, idputil.DischargeID(req), err)
		}
	}
}

func (idp *identityProvider) handleLogin(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	switch req.Method {
	default:
		return errgo.WithCausef(nil, params.ErrBadRequest, "unsupported method %q", req.Method)
	case "GET":
		return errgo.Mask(idp.initParams.Template.ExecuteTemplate(w, "login-form", nil))
	case "POST":
		id, err := idp.loginUser(ctx, req.Form.Get("username"), req.Form.Get("password"))
		if err != nil {
			return err
		}
		idp.initParams.VisitCompleter.Success(ctx, w, req, idputil.DischargeID(req), id)
		return nil
	}
}

func (idp *identityProvider) loginUser(ctx context.Context, user, password string) (*store.Identity, error) {
	if userData, ok := idp.params.Users[user]; ok {
		if userData.Password == password {
			username := idputil.NameWithDomain(user, idp.params.Domain)
			id := &store.Identity{
				ProviderID: store.MakeProviderIdentity(idp.params.Name, username),
				Username:   username,
				Name:       userData.Name,
				Email:      userData.Email,
			}
			err := idp.initParams.Store.UpdateIdentity(ctx, id, store.Update{
				store.Username: store.Set,
				store.Name:     store.Set,
				store.Email:    store.Set,
			})
			if err != nil {
				return nil, errgo.Mask(err)
			}
			return id, nil
		}
	}
	return nil, errgo.Newf("authentication failed for user %q", user)
}
