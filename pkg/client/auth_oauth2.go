// Copyright © 2020 The Things Industries B.V.

package client

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/credentials"
)

// DefaultTokenURL is the default Packet Broker IAM token URL.
const DefaultTokenURL = "https://iam.packetbroker.org/token"

type clientCredentials struct {
	tokenSource oauth2.TokenSource
	accessToken,
	refreshToken string
	insecure bool
}

func (c *clientCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("client: get token: %w", err)
	}
	if !token.Valid() {
		return nil, fmt.Errorf("client: invalid token")
	}
	return map[string]string{
		"authorization": fmt.Sprintf("%s %s", token.Type(), token.AccessToken),
	}, nil
}

func (c *clientCredentials) RequireTransportSecurity() bool {
	return !c.insecure
}

// OAuth2 returns per RPC client credentials using the OAuth Client Credentials flow.
// The token is being refreshed in the background.
func OAuth2(ctx context.Context, tokenURL, clientID, clientSecret string, scopes []string, insecure bool) credentials.PerRPCCredentials {
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		AuthStyle:    oauth2.AuthStyleInParams,
		TokenURL:     tokenURL,
	}
	return &clientCredentials{
		// TODO: Cache tokens on disk (https://github.com/packetbroker/pb/issues/6)
		tokenSource: config.TokenSource(ctx),
		insecure:    insecure,
	}
}
