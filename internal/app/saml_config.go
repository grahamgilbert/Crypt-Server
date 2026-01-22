package app

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type SAMLConfig struct {
	RootURL             string            `yaml:"root_url"`
	EntityID            string            `yaml:"entity_id"`
	IDPMetadataPath     string            `yaml:"idp_metadata_path"`
	IDPMetadataURL      string            `yaml:"idp_metadata_url"`
	CertificatePath     string            `yaml:"certificate_path"`
	PrivateKeyPath      string            `yaml:"private_key_path"`
	AllowIDPInitiated   bool              `yaml:"allow_idp_initiated"`
	SignRequest         bool              `yaml:"sign_request"`
	UseNameIDAsUsername bool              `yaml:"use_name_id_as_username"`
	CreateUnknownUser   bool              `yaml:"create_unknown_user"`
	UsernameAttribute   string            `yaml:"username_attribute"`
	AttributeMapping    map[string]string `yaml:"attribute_mapping"`
	GroupsAttribute     string            `yaml:"groups_attribute"`
	StaffGroups         []string          `yaml:"staff_groups"`
	SuperuserGroups     []string          `yaml:"superuser_groups"`
	CanApproveGroups    []string          `yaml:"can_approve_groups"`
	DefaultAuthSource   string            `yaml:"auth_source"`
	DefaultLocalLogin   bool              `yaml:"local_login_enabled"`
	DefaultMustReset    bool              `yaml:"must_reset_password"`
	DefaultRedirectURI  string            `yaml:"default_redirect_uri"`
	MetadataURLPath     string            `yaml:"metadata_url_path"`
	AcsURLPath          string            `yaml:"acs_url_path"`
	SloURLPath          string            `yaml:"slo_url_path"`
}

func LoadSAMLConfig(path string) (*SAMLConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read saml config: %w", err)
	}
	var cfg SAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse saml yaml: %w", err)
	}
	if cfg.RootURL == "" {
		return nil, errors.New("saml config missing root_url")
	}
	if cfg.IDPMetadataPath == "" && cfg.IDPMetadataURL == "" {
		return nil, errors.New("saml config missing idp metadata path or url")
	}
	// Certificate and private key are optional - only needed if sign_request is true
	// or if the IdP encrypts assertions
	if cfg.SignRequest && (cfg.CertificatePath == "" || cfg.PrivateKeyPath == "") {
		return nil, errors.New("saml config requires certificate and private key when sign_request is enabled")
	}
	if cfg.GroupsAttribute == "" {
		cfg.GroupsAttribute = "memberOf"
	}
	if cfg.DefaultAuthSource == "" {
		cfg.DefaultAuthSource = "saml"
	}
	if cfg.DefaultRedirectURI == "" {
		cfg.DefaultRedirectURI = "/"
	}
	if cfg.MetadataURLPath == "" {
		cfg.MetadataURLPath = "/saml2/metadata/"
	}
	if cfg.AcsURLPath == "" {
		cfg.AcsURLPath = "/saml2/acs/"
	}
	if cfg.SloURLPath == "" {
		cfg.SloURLPath = "/saml2/ls/"
	}
	return &cfg, nil
}
