// Copyright © 2020 The Things Industries B.V.

package client

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc/credentials"
)

type basicAuth struct {
	username, password string
	insecure           bool
}

func (b *basicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	authValue := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.password)))
	return map[string]string{
		"authorization": fmt.Sprintf("Basic %s", authValue),
	}, nil
}

func (b *basicAuth) RequireTransportSecurity() bool {
	return !b.insecure
}

// BasicAuth returns per RPC client credentials using basic authentication.
func BasicAuth(username, password string, insecure bool) credentials.PerRPCCredentials {
	return &basicAuth{
		username: username,
		password: password,
		insecure: insecure,
	}
}
