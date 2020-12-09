// Copyright © 2020 The Things Industries B.V.

// Package config provides configuration used by commands.
package config

import (
	"context"
	"errors"
	"os"

	flag "github.com/spf13/pflag"
	"go.packetbroker.org/pb/pkg/client"
)

// CommonFlags defines common flags.
func CommonFlags(help, debug *bool) {
	flag.BoolVarP(help, "help", "h", false, "show usage")
	flag.BoolVarP(debug, "debug", "d", false, "debug mode")
}

// ClientFlags defines common flags used for Client configuration.
func ClientFlags(defaultAddress string) {
	if defaultAddress != "" {
		flag.String("address", "", `address of the server "host[:port]" (default $PB_ADDRESS)`)
	} else {
		flag.String("address", defaultAddress, `address of the server "host[:port]"`)
	}
	flag.Bool("insecure", false, "insecure")
}

// BasicAuthClientFlags defines flags used for Basic authentication.
func BasicAuthClientFlags() {
	flag.StringP("username", "u", "", "IAM username (default $PB_IAM_USERNAME)")
	flag.StringP("password", "p", "", "IAM password (default $PB_IAM_PASSWORD)")
}

// OAuth2ClientFlags defines flags used for OAuth Client Credentials flow.
func OAuth2ClientFlags() {
	flag.String("client-id", "", "OAuth 2.0 client ID (default $PB_CLIENT_ID)")
	flag.String("client-secret", "", "OAuth 2.0 client secret (default $PB_CLIENT_SECRET)")
	flag.String("token-url", client.DefaultTokenURL, "OAuth 2.0 token URL")
}

// initClient returns initial client configuration.
func initClient() (*client.Config, error) {
	var (
		res client.Config
		err error
	)
	if res.Address, err = flag.CommandLine.GetString("address"); err != nil {
		return nil, err
	}
	if res.Address == "" {
		res.Address = os.Getenv("PB_ADDRESS")
	}
	if res.Address == "" {
		return nil, errors.New("missing server address")
	}
	res.Insecure, _ = flag.CommandLine.GetBool("insecure")
	return &res, nil
}

var errNoCredentials = errors.New("no credentials")

// BasicAuthClient returns a client configured with Basic authentication.
func BasicAuthClient() (*client.Config, error) {
	res, err := initClient()
	if err != nil {
		return nil, err
	}
	var username, password string
	if username, err = flag.CommandLine.GetString("username"); err != nil {
		return nil, err
	}
	if username == "" {
		username = os.Getenv("PB_IAM_USERNAME")
	}
	if password, err = flag.CommandLine.GetString("password"); err != nil {
		return nil, err
	}
	if password == "" {
		password = os.Getenv("PB_IAM_PASSWORD")
	}
	if username == "" || password == "" {
		return nil, errNoCredentials
	}
	allowInsecure, _ := flag.CommandLine.GetBool("insecure")
	res.Credentials = client.BasicAuth(username, password, allowInsecure)
	return res, nil
}

// OAuth2Client returns a client configured with OAuth Client Credentials authentication.
func OAuth2Client(ctx context.Context, scopes ...string) (*client.Config, error) {
	res, err := initClient()
	if err != nil {
		return nil, err
	}
	var tokenURL, clientID, clientSecret string
	if tokenURL, err = flag.CommandLine.GetString("token-url"); err != nil {
		return nil, err
	}
	if clientID, err = flag.CommandLine.GetString("client-id"); err != nil {
		return nil, err
	}
	if clientID == "" {
		clientID = os.Getenv("PB_CLIENT_ID")
	}
	if clientSecret, err = flag.CommandLine.GetString("client-secret"); err != nil {
		return nil, err
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("PB_CLIENT_SECRET")
	}
	if clientID == "" || clientSecret == "" {
		return nil, errNoCredentials
	}
	allowInsecure, _ := flag.CommandLine.GetBool("insecure")
	res.Credentials = client.OAuth2(ctx, tokenURL, clientID, clientSecret, scopes, allowInsecure)
	return res, nil
}

// AutomaticClient returns a client configured based on available settings.
// Basic authentication is preferred. Otherwise, OAuth Client Credentials are used.
func AutomaticClient(ctx context.Context, oauthScopes ...string) (*client.Config, error) {
	config, err := BasicAuthClient()
	if err == nil {
		return config, nil
	}
	if errors.Is(err, errNoCredentials) {
		return OAuth2Client(ctx, oauthScopes...)
	}
	return nil, err
}
