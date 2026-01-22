package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSAMLConfigDefaults(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "saml.yaml")
	err := os.WriteFile(cfgPath, []byte(`root_url: https://crypt.example.com
idp_metadata_path: /tmp/metadata.xml
`), 0o600)
	require.NoError(t, err)

	cfg, err := LoadSAMLConfig(cfgPath)
	require.NoError(t, err)
	require.Equal(t, "/saml2/metadata/", cfg.MetadataURLPath)
	require.Equal(t, "/saml2/acs/", cfg.AcsURLPath)
	require.Equal(t, "/saml2/ls/", cfg.SloURLPath)
	require.Equal(t, "memberOf", cfg.GroupsAttribute)
	require.Equal(t, "saml", cfg.DefaultAuthSource)
	require.Equal(t, "/", cfg.DefaultRedirectURI)
}

func TestLoadSAMLConfigWithoutCertificates(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "saml.yaml")
	// Config without certificate/private_key - should work when sign_request is false
	err := os.WriteFile(cfgPath, []byte(`root_url: https://crypt.example.com
idp_metadata_url: https://idp.example.com/metadata
`), 0o600)
	require.NoError(t, err)

	cfg, err := LoadSAMLConfig(cfgPath)
	require.NoError(t, err)
	require.Empty(t, cfg.CertificatePath)
	require.Empty(t, cfg.PrivateKeyPath)
}

func TestLoadSAMLConfigSignRequestRequiresCertificates(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "saml.yaml")
	// sign_request: true requires certificate and private key
	err := os.WriteFile(cfgPath, []byte(`root_url: https://crypt.example.com
idp_metadata_url: https://idp.example.com/metadata
sign_request: true
`), 0o600)
	require.NoError(t, err)

	_, err = LoadSAMLConfig(cfgPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "certificate and private key")
}

func TestLoadSAMLConfigRequiresFields(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "saml.yaml")
	err := os.WriteFile(cfgPath, []byte(`root_url: ""`), 0o600)
	require.NoError(t, err)

	_, err = LoadSAMLConfig(cfgPath)
	require.Error(t, err)
}
