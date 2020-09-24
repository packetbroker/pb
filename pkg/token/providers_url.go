// Copyright © 2020 The Things Industries B.V.

package token

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/square/go-jose.v2"
)

type urlPublicKeyProvider struct {
	url string
}

func (p *urlPublicKeyProvider) PublicKeys() (*jose.JSONWebKeySet, error) {
	res, err := http.Get(p.url)
	if err != nil {
		return nil, fmt.Errorf("token: fetch from %q: %w", p.url, err)
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("token: read response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("token: unsuccessful response: %s", res.Status)
	}
	key := new(jose.JSONWebKeySet)
	if err := json.Unmarshal(buf, key); err != nil {
		return nil, fmt.Errorf("token: parse response: %w", err)
	}
	return key, nil
}

// PublicKeyFromURL returns a PublicKeyProvider that reads the given JSON Web Key Set from a URL.
func PublicKeyFromURL(url string) PublicKeyProvider {
	return &filePublicKeyProvider{url}
}
