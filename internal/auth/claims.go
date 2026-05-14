package auth

import "github.com/golang-jwt/jwt/v5"

type UserPrincipal struct {
	UserID   string
	Username string
	Email    string
	Roles    []string
}

const (
	RoleRequestor = "Requestor"
	RoleAgent     = "Agent"
	RoleManager   = "Manager"
)

type RealmAccess struct {
	Roles []string `json:"roles"`
}

type KeycloakClaims struct {
	jwt.RegisteredClaims

	AuthorizedParty   string      `json:"azp"`
	PreferredUsername string      `json:"preferred_username"`
	Email             string      `json:"email"`
	RealmAccess       RealmAccess `json:"realm_access"`
}

func (c KeycloakClaims) ToPrincipal() UserPrincipal {
	return UserPrincipal{
		UserID:   c.Subject,
		Username: c.PreferredUsername,
		Email:    c.Email,
		Roles:    c.RealmAccess.Roles,
	}
}

func (u UserPrincipal) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u UserPrincipal) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

func (u UserPrincipal) BusinessRoles() []string {
	var roles []string

	for _, role := range u.Roles {
		switch role {
		case RoleRequestor, RoleAgent, RoleManager:
			roles = append(roles, role)
		}
	}

	return roles
}

func (u UserPrincipal) HasBusinessRole() bool {
	return len(u.BusinessRoles()) > 0
}
