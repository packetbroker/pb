// Copyright © 2020 The Things Industries B.V.

package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/credentials"
)

type clientCredential struct {
	authType,
	authValue string
	authMu   sync.RWMutex
	authDone chan struct{}
	insecure bool
}

// UpdateAuthValue updates the authorization value.
func (c *clientCredential) UpdateAuthValue(authValue string) {
	c.authMu.Lock()
	c.authValue = authValue
	c.authMu.Unlock()
}

func (c *clientCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if c.authDone != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.authDone:
		}
	}
	c.authMu.RLock()
	authType, authValue := c.authType, c.authValue
	c.authMu.RUnlock()
	return map[string]string{
		"authorization": fmt.Sprintf("%s %s", authType, authValue),
	}, nil
}

func (c *clientCredential) RequireTransportSecurity() bool {
	return !c.insecure
}

// BasicAuth returns per RPC client credentials using basic authentication.
func BasicAuth(username, password string, insecure bool) credentials.PerRPCCredentials {
	return &clientCredential{
		authType:  "basic",
		authValue: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))),
		insecure:  insecure,
	}
}

// OAuthClient returns per RPC client credentials using the OAuth Client Credentials flow.
// The token is being refreshed in the background until the given context closes.
func OAuthClient(ctx context.Context, clientID, clientSecret string, insecure bool) credentials.PerRPCCredentials {
	cred := &clientCredential{
		authDone: make(chan struct{}),
		insecure: insecure,
	}
	go func() {
		for {
			// TODO: Get and refresh token
			validFor := time.Duration(1 * time.Hour)
			close(cred.authDone)

			select {
			case <-ctx.Done():
				return
			case <-time.After(validFor / 4 * 3):
			}
		}
	}()
	return cred
}
