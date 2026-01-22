package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]interface{}
		key      string
		expected string
	}{
		{"string value", map[string]interface{}{"foo": "bar"}, "foo", "bar"},
		{"missing key", map[string]interface{}{}, "foo", ""},
		{"nil value", map[string]interface{}{"foo": nil}, "foo", ""},
		{"int value", map[string]interface{}{"foo": 42}, "foo", "42"},
		{"float value", map[string]interface{}{"foo": 3.14}, "foo", "3.14"},
		{"bool value", map[string]interface{}{"foo": true}, "foo", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, getString(tt.fields, tt.key))
		})
	}
}

func TestGetInt(t *testing.T) {
	jsonNum := json.Number("42")
	tests := []struct {
		name     string
		fields   map[string]interface{}
		key      string
		expected int
	}{
		{"float64 value", map[string]interface{}{"foo": float64(42)}, "foo", 42},
		{"int value", map[string]interface{}{"foo": 42}, "foo", 42},
		{"json.Number value", map[string]interface{}{"foo": jsonNum}, "foo", 42},
		{"missing key", map[string]interface{}{}, "foo", 0},
		{"nil value", map[string]interface{}{"foo": nil}, "foo", 0},
		{"string value", map[string]interface{}{"foo": "bar"}, "foo", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, getInt(tt.fields, tt.key))
		})
	}
}

func TestGetOptionalInt(t *testing.T) {
	jsonNum := json.Number("42")
	t.Run("float64 value", func(t *testing.T) {
		result := getOptionalInt(map[string]interface{}{"foo": float64(42)}, "foo")
		require.NotNil(t, result)
		require.Equal(t, 42, *result)
	})

	t.Run("int value", func(t *testing.T) {
		result := getOptionalInt(map[string]interface{}{"foo": 42}, "foo")
		require.NotNil(t, result)
		require.Equal(t, 42, *result)
	})

	t.Run("json.Number value", func(t *testing.T) {
		result := getOptionalInt(map[string]interface{}{"foo": jsonNum}, "foo")
		require.NotNil(t, result)
		require.Equal(t, 42, *result)
	})

	t.Run("missing key", func(t *testing.T) {
		result := getOptionalInt(map[string]interface{}{}, "foo")
		require.Nil(t, result)
	})

	t.Run("nil value", func(t *testing.T) {
		result := getOptionalInt(map[string]interface{}{"foo": nil}, "foo")
		require.Nil(t, result)
	})

	t.Run("string value", func(t *testing.T) {
		result := getOptionalInt(map[string]interface{}{"foo": "bar"}, "foo")
		require.Nil(t, result)
	})
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]interface{}
		key      string
		expected bool
	}{
		{"true value", map[string]interface{}{"foo": true}, "foo", true},
		{"false value", map[string]interface{}{"foo": false}, "foo", false},
		{"missing key", map[string]interface{}{}, "foo", false},
		{"nil value", map[string]interface{}{"foo": nil}, "foo", false},
		{"string value", map[string]interface{}{"foo": "true"}, "foo", false},
		{"int value", map[string]interface{}{"foo": 1}, "foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, getBool(tt.fields, tt.key))
		})
	}
}

func TestGetOptionalBool(t *testing.T) {
	t.Run("true value", func(t *testing.T) {
		result := getOptionalBool(map[string]interface{}{"foo": true}, "foo")
		require.NotNil(t, result)
		require.True(t, *result)
	})

	t.Run("false value", func(t *testing.T) {
		result := getOptionalBool(map[string]interface{}{"foo": false}, "foo")
		require.NotNil(t, result)
		require.False(t, *result)
	})

	t.Run("missing key", func(t *testing.T) {
		result := getOptionalBool(map[string]interface{}{}, "foo")
		require.Nil(t, result)
	})

	t.Run("nil value", func(t *testing.T) {
		result := getOptionalBool(map[string]interface{}{"foo": nil}, "foo")
		require.Nil(t, result)
	})

	t.Run("string value", func(t *testing.T) {
		result := getOptionalBool(map[string]interface{}{"foo": "true"}, "foo")
		require.Nil(t, result)
	})
}

func TestUsernameForID(t *testing.T) {
	users := map[int]string{
		1: "admin",
		2: "user",
	}

	t.Run("existing user", func(t *testing.T) {
		id := 1
		require.Equal(t, "admin", usernameForID(users, &id))
	})

	t.Run("nil id", func(t *testing.T) {
		require.Equal(t, "", usernameForID(users, nil))
	})

	t.Run("missing user", func(t *testing.T) {
		id := 999
		require.Equal(t, "user-999", usernameForID(users, &id))
	})
}

func TestMapGroups(t *testing.T) {
	groups := map[int]string{
		1: "admins",
		2: "approvers",
		3: "users",
	}

	t.Run("multiple groups", func(t *testing.T) {
		result := mapGroups([]int{1, 2}, groups)
		require.Equal(t, []string{"admins", "approvers"}, result)
	})

	t.Run("empty group ids", func(t *testing.T) {
		result := mapGroups([]int{}, groups)
		require.Empty(t, result)
	})

	t.Run("missing group", func(t *testing.T) {
		result := mapGroups([]int{1, 999}, groups)
		require.Equal(t, []string{"admins"}, result)
	})

	t.Run("duplicate groups", func(t *testing.T) {
		result := mapGroups([]int{1, 1, 2}, groups)
		require.Equal(t, []string{"admins", "approvers"}, result)
	})
}

func TestHasPermission(t *testing.T) {
	allowed := map[int]struct{}{
		10: {},
		20: {},
	}

	t.Run("has permission", func(t *testing.T) {
		require.True(t, hasPermission([]int{10}, allowed))
		require.True(t, hasPermission([]int{5, 10, 15}, allowed))
	})

	t.Run("no permission", func(t *testing.T) {
		require.False(t, hasPermission([]int{5, 15}, allowed))
		require.False(t, hasPermission([]int{}, allowed))
	})
}

func TestResolveCanApprove(t *testing.T) {
	canApproveIDs := map[int]struct{}{100: {}}

	t.Run("superuser can approve", func(t *testing.T) {
		user := userOut{IsSuper: true}
		require.True(t, resolveCanApprove(user, nil, nil, nil, canApproveIDs))
	})

	t.Run("user with direct permission", func(t *testing.T) {
		user := userOut{}
		require.True(t, resolveCanApprove(user, []int{100}, nil, nil, canApproveIDs))
	})

	t.Run("user with group permission", func(t *testing.T) {
		user := userOut{}
		groupPerms := map[int][]int{
			5: {100},
		}
		require.True(t, resolveCanApprove(user, nil, groupPerms, []int{5}, canApproveIDs))
	})

	t.Run("user without permission", func(t *testing.T) {
		user := userOut{}
		require.False(t, resolveCanApprove(user, []int{50}, nil, nil, canApproveIDs))
	})
}

func TestApplyPasswordMapping(t *testing.T) {
	passwordMap := map[string]passwordMapEntry{
		"admin": {PasswordHash: "hash1", MustResetPassword: false},
		"user@example.com": {PasswordHash: "hash2", MustResetPassword: true},
	}

	t.Run("match by username", func(t *testing.T) {
		user := &userOut{Username: "admin"}
		applyPasswordMapping(user, passwordMap, "")
		require.Equal(t, "hash1", user.PasswordHash)
		require.False(t, user.MustResetPassword)
		require.True(t, user.LocalLoginEnabled)
		require.Equal(t, "local", user.AuthSource)
	})

	t.Run("match by email", func(t *testing.T) {
		user := &userOut{Username: "someone"}
		applyPasswordMapping(user, passwordMap, "user@example.com")
		require.Equal(t, "hash2", user.PasswordHash)
		require.True(t, user.MustResetPassword)
		require.True(t, user.LocalLoginEnabled)
		require.Equal(t, "local", user.AuthSource)
	})

	t.Run("no match defaults to saml", func(t *testing.T) {
		user := &userOut{Username: "unknown"}
		applyPasswordMapping(user, passwordMap, "unknown@example.com")
		require.Empty(t, user.PasswordHash)
		require.True(t, user.MustResetPassword)
		require.False(t, user.LocalLoginEnabled)
		require.Equal(t, "saml", user.AuthSource)
	})

	t.Run("case insensitive username match", func(t *testing.T) {
		user := &userOut{Username: "Admin"}
		applyPasswordMapping(user, passwordMap, "")
		// Note: the current implementation lowercases username for lookup
		require.Equal(t, "hash1", user.PasswordHash)
	})
}

func TestMarshalOutput(t *testing.T) {
	output := &migrationOutput{
		Computers: []computerOut{{ID: 1, Serial: "ABC"}},
		Secrets:   []secretOut{},
		Requests:  []requestOut{},
		Users:     []userOut{},
	}

	data, err := marshalOutput(output)
	require.NoError(t, err)
	require.Contains(t, string(data), `"serial": "ABC"`)
	require.Contains(t, string(data), `"id": 1`)
}

func TestDecryptLegacySecretEmpty(t *testing.T) {
	_, err := decryptLegacySecret("", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty secret")
}
