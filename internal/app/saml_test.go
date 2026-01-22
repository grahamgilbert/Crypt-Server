package app

import (
	"testing"

	"github.com/crewjam/saml/samlsp"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestUsernameFromSAMLNameID(t *testing.T) {
	cfg := &SAMLConfig{UseNameIDAsUsername: true}
	claims := samlsp.JWTSessionClaims{StandardClaims: jwt.StandardClaims{Subject: "nameid-user"}}
	require.Equal(t, "nameid-user", usernameFromSAML(claims, cfg))
}

func TestUsernameFromSAMLAttributeMapping(t *testing.T) {
	cfg := &SAMLConfig{
		UseNameIDAsUsername: false,
		AttributeMapping:    map[string]string{"uid": "username"},
	}
	claims := samlsp.JWTSessionClaims{
		Attributes: samlsp.Attributes{
			"uid": []string{"mapped-user"},
		},
	}
	require.Equal(t, "mapped-user", usernameFromSAML(claims, cfg))
}

func TestUsernameFromSAMLUsernameAttribute(t *testing.T) {
	cfg := &SAMLConfig{
		UseNameIDAsUsername: false,
		UsernameAttribute:   "email",
	}
	claims := samlsp.JWTSessionClaims{
		Attributes: samlsp.Attributes{
			"email": []string{"user@example.com"},
		},
	}
	require.Equal(t, "user@example.com", usernameFromSAML(claims, cfg))
}

func TestResolveSAMLPermissions(t *testing.T) {
	cfg := &SAMLConfig{
		StaffGroups:      []string{"staff"},
		SuperuserGroups:  []string{"super"},
		CanApproveGroups: []string{"approvers"},
	}
	isStaff, canApprove := resolveSAMLPermissions([]string{"approvers"}, cfg)
	require.False(t, isStaff)
	require.True(t, canApprove)

	isStaff, canApprove = resolveSAMLPermissions([]string{"super"}, cfg)
	require.True(t, isStaff)
	require.True(t, canApprove)
}
