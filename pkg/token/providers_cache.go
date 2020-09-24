// Copyright © 2020 The Things Industries B.V.

package token

import (
	"sync"

	"gopkg.in/square/go-jose.v2"
)

// CachePublicKey caches the result from the given PublicKeyProvider indefinitely.
func CachePublicKey(provider PublicKeyProvider) PublicKeyProvider {
	var (
		key *jose.JSONWebKeySet
		err error
		mu  sync.Mutex
	)
	return PublicKeyProviderFunc(func() (*jose.JSONWebKeySet, error) {
		mu.Lock()
		defer mu.Unlock()
		if key == nil && err == nil {
			key, err = provider.PublicKeys()
		}
		return key, err
	})
}

// CachePrivateKey caches the result from the given PrivateKeyProvider indefinitely.
func CachePrivateKey(provider PrivateKeyProvider) PrivateKeyProvider {
	var (
		key *jose.JSONWebKey
		err error
		mu  sync.Mutex
	)
	return PrivateKeyProviderFunc(func() (*jose.JSONWebKey, error) {
		mu.Lock()
		defer mu.Unlock()
		if key == nil && err == nil {
			key, err = provider.PrivateKey()
		}
		return key, err
	})
}
