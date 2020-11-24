// Copyright © 2020 The Things Industries B.V.

package token

import (
	"fmt"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// PublicKeyProvider provides a set of public keys.
type PublicKeyProvider interface {
	PublicKeys() (*jose.JSONWebKeySet, error)
}

// PublicKeyProviderFunc is a function that implements PublicKeyProvider.
type PublicKeyProviderFunc func() (*jose.JSONWebKeySet, error)

// PublicKeys implements PublicKeyProvider.
func (f PublicKeyProviderFunc) PublicKeys() (*jose.JSONWebKeySet, error) {
	return f()
}

// PrivateKeyProvider provides a private key.
type PrivateKeyProvider interface {
	PrivateKey() (*jose.JSONWebKey, error)
}

// PrivateKeyProviderFunc is a function that implements PrivateKeyProvider.
type PrivateKeyProviderFunc func() (*jose.JSONWebKey, error)

// PrivateKey implements PrivateKeyProvider.
func (f PrivateKeyProviderFunc) PrivateKey() (*jose.JSONWebKey, error) {
	return f()
}

// Network defines a Packet Broker network.
type Network struct {
	NetID    uint32 `json:"nid"`
	TenantID string `json:"tid,omitempty"`
	ID       string `json:"id,omitempty"`
}

// PacketBrokerClaims defines claims specific for Packet Broker.
type PacketBrokerClaims struct {
	Networks []Network `json:"ns,omitempty"`
}

// Claims defines the JSON Web Token claims.
type Claims struct {
	jwt.Claims
	PacketBroker *PacketBrokerClaims `json:"https://iam.packetbroker.org/claims,omitempty"`
}

// Sign signs the claims with the given key and returns the compact token.
func Sign(keyProvider PrivateKeyProvider, claims Claims) (string, error) {
	key, err := keyProvider.PrivateKey()
	if err != nil {
		return "", fmt.Errorf("token: private key: %w", err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.SignatureAlgorithm(key.Algorithm),
		Key:       key,
	}, new(jose.SignerOptions).WithType("JWT"))
	if err != nil {
		return "", fmt.Errorf("token: instantiate signer: %w", err)
	}
	token, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("token: sign: %w", err)
	}
	return token, nil
}

// Parse parses and verifies the token and returns the claims.
func Parse(keyProvider PublicKeyProvider, token string) (Claims, error) {
	t, err := jwt.ParseSigned(token)
	if err != nil {
		return Claims{}, fmt.Errorf("token: parse: %w", err)
	}
	keys, err := keyProvider.PublicKeys()
	if err != nil {
		return Claims{}, fmt.Errorf("token: public keys: %w", err)
	}
	var claims Claims
	if err := t.Claims(keys, &claims); err != nil {
		return Claims{}, fmt.Errorf("token: verify: %w", err)
	}
	return claims, nil
}
