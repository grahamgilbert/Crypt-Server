package app

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

func BuildSAMLProvider(cfg *SAMLConfig) (*samlsp.Middleware, error) {
	rootURL, err := url.Parse(cfg.RootURL)
	if err != nil {
		return nil, fmt.Errorf("parse saml root url: %w", err)
	}

	idpMetadata, err := loadIDPMetadata(cfg)
	if err != nil {
		return nil, err
	}

	entityID := cfg.EntityID
	if entityID == "" {
		entityID = rootURL.ResolveReference(&url.URL{Path: cfg.MetadataURLPath}).String()
	}

	opts := samlsp.Options{
		EntityID:           entityID,
		URL:                *rootURL,
		IDPMetadata:        idpMetadata,
		AllowIDPInitiated:  cfg.AllowIDPInitiated,
		DefaultRedirectURI: cfg.DefaultRedirectURI,
		SignRequest:        cfg.SignRequest,
	}

	// Load certificate and private key if provided
	if cfg.CertificatePath != "" && cfg.PrivateKeyPath != "" {
		keyPair, err := tls.LoadX509KeyPair(cfg.CertificatePath, cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load saml keypair: %w", err)
		}

		cert, err := parseX509Certificate(keyPair.Certificate)
		if err != nil {
			return nil, fmt.Errorf("parse saml certificate: %w", err)
		}

		privateKey, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("saml private key must be RSA")
		}

		opts.Key = privateKey
		opts.Certificate = cert
	}

	middleware, err := samlsp.New(opts)
	if err != nil {
		return nil, fmt.Errorf("init saml: %w", err)
	}

	metadataURL := rootURL.ResolveReference(&url.URL{Path: cfg.MetadataURLPath})
	acsURL := rootURL.ResolveReference(&url.URL{Path: cfg.AcsURLPath})
	sloURL := rootURL.ResolveReference(&url.URL{Path: cfg.SloURLPath})
	middleware.ServiceProvider.MetadataURL = *metadataURL
	middleware.ServiceProvider.AcsURL = *acsURL
	middleware.ServiceProvider.SloURL = *sloURL

	return middleware, nil
}

func loadIDPMetadata(cfg *SAMLConfig) (*saml.EntityDescriptor, error) {
	if cfg.IDPMetadataPath != "" {
		data, err := os.ReadFile(cfg.IDPMetadataPath)
		if err != nil {
			return nil, fmt.Errorf("read idp metadata: %w", err)
		}
		metadata, err := samlsp.ParseMetadata(data)
		if err != nil {
			return nil, fmt.Errorf("parse idp metadata: %w", err)
		}
		return metadata, nil
	}
	metadataURL, err := url.Parse(cfg.IDPMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("parse idp metadata url: %w", err)
	}
	metadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *metadataURL)
	if err != nil {
		return nil, fmt.Errorf("fetch idp metadata: %w", err)
	}
	return metadata, nil
}

func parseX509Certificate(certs [][]byte) (*x509.Certificate, error) {
	if len(certs) == 0 {
		return nil, errors.New("missing certificate")
	}
	return x509.ParseCertificate(certs[0])
}

func usernameFromSAML(session samlsp.Session, cfg *SAMLConfig) string {
	if cfg.UseNameIDAsUsername {
		if claims, ok := session.(samlsp.JWTSessionClaims); ok {
			if claims.Subject != "" {
				return claims.Subject
			}
		}
	}
	attributes := attributesFromSession(session)
	if cfg.UsernameAttribute != "" {
		if value := attributes.Get(cfg.UsernameAttribute); value != "" {
			return value
		}
	}
	for key, mapped := range cfg.AttributeMapping {
		if mapped == "username" {
			if value := attributes.Get(key); value != "" {
				return value
			}
		}
	}
	if value := attributes.Get("uid"); value != "" {
		return value
	}
	return ""
}

func attributesFromSession(session samlsp.Session) samlsp.Attributes {
	if session == nil {
		return nil
	}
	if withAttrs, ok := session.(samlsp.SessionWithAttributes); ok {
		return withAttrs.GetAttributes()
	}
	return nil
}

func groupMembership(attributes samlsp.Attributes, attr string) []string {
	if attributes == nil {
		return nil
	}
	return attributes[attr]
}

func matchesGroup(groups []string, target []string) bool {
	for _, candidate := range groups {
		for _, group := range target {
			if strings.EqualFold(candidate, group) {
				return true
			}
		}
	}
	return false
}

func resolveSAMLPermissions(groups []string, cfg *SAMLConfig) (bool, bool) {
	isStaff := false
	canApprove := false
	if matchesGroup(groups, cfg.SuperuserGroups) {
		isStaff = true
		canApprove = true
	}
	if matchesGroup(groups, cfg.StaffGroups) {
		isStaff = true
	}
	if matchesGroup(groups, cfg.CanApproveGroups) {
		canApprove = true
	}
	return isStaff, canApprove
}
