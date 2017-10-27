package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/influxdata/chronograf"
	"github.com/influxdata/chronograf/oauth2"
)

// AuthorizedToken extracts the token and validates; if valid the next handler
// will be run.  The principal will be sent to the next handler via the request's
// Context.  It is up to the next handler to determine if the principal has access.
// On failure, will return http.StatusForbidden.
func AuthorizedToken(auth oauth2.Authenticator, logger chronograf.Logger, next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.
			WithField("component", "token_auth").
			WithField("remote_addr", r.RemoteAddr).
			WithField("method", r.Method).
			WithField("url", r.URL)

		ctx := r.Context()
		// We do not check the authorization of the principal.  Those
		// served further down the chain should do so.
		principal, err := auth.Validate(ctx, r)
		if err != nil {
			log.Error("Invalid principal")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// If the principal is valid we will extend its lifespan
		// into the future
		principal, err = auth.Extend(ctx, w, principal)
		if err != nil {
			log.Error("Unable to extend principal")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Send the principal to the next handler
		ctx = context.WithValue(ctx, oauth2.PrincipalKey, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

// AuthorizedUser extracts the user name and provider from context. If the
// user and provider can be found on the context, we look up the user by their
// name and provider. If the user is found, we verify that the user has at at
// least the role supplied.
func AuthorizedUser(
	usersStore chronograf.UsersStore,
	organizationsStore chronograf.OrganizationsStore,
	useAuth bool,
	role string,
	logger chronograf.Logger,
	next http.HandlerFunc,
) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !useAuth {
			next(w, r)
			return
		}

		log := logger.
			WithField("component", "role_auth").
			WithField("remote_addr", r.RemoteAddr).
			WithField("method", r.Method).
			WithField("url", r.URL)

		ctx := r.Context()

		p, err := getValidPrincipal(ctx)
		if err != nil {
			log.Error("Failed to retrieve principal from context")
			Error(w, http.StatusUnauthorized, "User is not authorized", logger)
			return
		}
		scheme, err := getScheme(ctx)
		if err != nil {
			log.Error("Failed to retrieve scheme from context")
			Error(w, http.StatusUnauthorized, "User is not authorized", logger)
			return
		}

		if p.Organization == "" {
			log.Error("Failed to retrieve organization from principal")
			Error(w, http.StatusUnauthorized, "User is not authorized", logger)
			return
		}
		// validate that the organization exists
		orgID, err := strconv.ParseUint(p.Organization, 10, 64)
		if err != nil {
			log.Error("Failed to validate organization on context")
			Error(w, http.StatusUnauthorized, "User is not authorized", logger)
			return
		}
		_, err = organizationsStore.Get(ctx, chronograf.OrganizationQuery{ID: &orgID})
		if err != nil {
			log.Error("Failed to retrieve organization from organizations store")
			Error(w, http.StatusUnauthorized, "User is not authorized", logger)
			return
		}

		ctx = context.WithValue(ctx, "organizationID", p.Organization)
		u, err := usersStore.Get(ctx, chronograf.UserQuery{
			Name:     &p.Subject,
			Provider: &p.Issuer,
			Scheme:   &scheme,
		})
		if err != nil {
			log.Error("Failed to retrieve user")
			Error(w, http.StatusUnauthorized, "User is not authorized", logger)
			return
		}

		if hasAuthorizedRole(u, role) {
			next(w, r)
			return
		}

		Error(w, http.StatusUnauthorized, "User is not authorized", logger)
		return
	})
}

func hasAuthorizedRole(u *chronograf.User, role string) bool {
	if u == nil {
		return false
	}

	switch role {
	case ViewerRoleName:
		for _, r := range u.Roles {
			switch r.Name {
			case ViewerRoleName, EditorRoleName, AdminRoleName:
				return true
			}
		}
	case EditorRoleName:
		for _, r := range u.Roles {
			switch r.Name {
			case EditorRoleName, AdminRoleName:
				return true
			}
		}
	case AdminRoleName:
		for _, r := range u.Roles {
			switch r.Name {
			case AdminRoleName:
				return true
			}
		}
	}

	return false
}
