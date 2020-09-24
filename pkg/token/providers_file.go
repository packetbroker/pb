// Copyright © 2020 The Things Industries B.V.

package token

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"gopkg.in/square/go-jose.v2"
)

type filePublicKeyProvider struct {
	path string
}

func (p *filePublicKeyProvider) PublicKeys() (*jose.JSONWebKeySet, error) {
	buf, err := ioutil.ReadFile(p.path)
	if err != nil {
		return nil, fmt.Errorf("token: read file %q: %w", p.path, err)
	}
	key := new(jose.JSONWebKeySet)
	if err := json.Unmarshal(buf, key); err != nil {
		return nil, fmt.Errorf("token: parse file %q: %w", p.path, err)
	}
	return key, nil
}

// PublicKeyFromFile returns a PublicKeyProvider that reads the given JSON Web Key Set file.
func PublicKeyFromFile(path string) PublicKeyProvider {
	return &filePublicKeyProvider{path}
}

type filePrivateKeyProvider struct {
	path string
}

func (p *filePrivateKeyProvider) PrivateKey() (*jose.JSONWebKey, error) {
	buf, err := ioutil.ReadFile(p.path)
	if err != nil {
		return nil, fmt.Errorf("token: read file %q: %w", p.path, err)
	}
	key := new(jose.JSONWebKey)
	if err := json.Unmarshal(buf, key); err != nil {
		return nil, fmt.Errorf("token: parse file %q: %w", p.path, err)
	}
	return key, nil
}

// PrivateKeyFromFile returns a PrivateKeyProvider thaat reads the given JSON Web Key file.
func PrivateKeyFromFile(path string) PrivateKeyProvider {
	return &filePrivateKeyProvider{path}
}
