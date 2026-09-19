// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration used by commands.
package config

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/packetbroker/pb/pkg/client"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BasicAuthRealm refers to a Basic authentication realm.
type BasicAuthRealm int

const (
	// BasicAuthIAM is the IAM realm.
	BasicAuthIAM BasicAuthRealm = iota
)

// BasicAuthRealmConfig defines the Basic authentication realm configuration.
type BasicAuthRealmConfig struct {
	ConfigKey,
	Name string
}

var basicAuthRealms = map[BasicAuthRealm]BasicAuthRealmConfig{
	BasicAuthIAM: {
		ConfigKey: "iam",
		Name:      "IAM",
	},
}

func mustRealmConfig(realm BasicAuthRealm) BasicAuthRealmConfig {
	conf, ok := basicAuthRealms[realm]
	if !ok {
		panic(fmt.Sprintf("realm %q not registered", rune(realm)))
	}
	return conf
}

// ClientFlags defines common flags used for Client configuration.
func ClientFlags(service, defaultAddress string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String(fmt.Sprintf("%s-address", service), defaultAddress, `address of the server "host[:port]"`)
	flags.Bool("insecure", false, "insecure")
	mustBindPFlags(flags)
	return flags
}

// mustBindPFlags binds the flags to viper. It panics if binding fails, which indicates a programming error.
func mustBindPFlags(flags *flag.FlagSet) {
	if err := viper.BindPFlags(flags); err != nil {
		panic(fmt.Sprintf("bind flags: %v", err))
	}
}

// BasicAuthClientFlags defines flags used for Basic authentication.
func BasicAuthClientFlags(realm BasicAuthRealm) *flag.FlagSet {
	flags := new(flag.FlagSet)
	conf := mustRealmConfig(realm)
	flags.String(fmt.Sprintf("%s-username", conf.ConfigKey), "", fmt.Sprintf("%s username", conf.Name))
	flags.String(fmt.Sprintf("%s-password", conf.ConfigKey), "", fmt.Sprintf("%s password", conf.Name))
	mustBindPFlags(flags)
	return flags
}

// OAuth2ClientFlags defines flags used for OAuth Client Credentials flow.
func OAuth2ClientFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String("client-id", "", "OAuth 2.0 client ID")
	flags.String("client-secret", "", "OAuth 2.0 client secret")
	flags.String("token-url", client.DefaultTokenURL, "OAuth 2.0 token URL")
	mustBindPFlags(flags)
	return flags
}

// initClient returns initial client configuration.
func initClient(service string) (*client.Config, error) {
	res := client.Config{
		Address:  viper.GetString(fmt.Sprintf("%s-address", service)),
		Insecure: viper.GetBool("insecure"),
	}
	if res.Address == "" {
		return nil, errors.New("missing server address")
	}
	return &res, nil
}

var errNoCredentials = errors.New("no credentials")

// BasicAuthClient returns a client configured with Basic authentication.
func BasicAuthClient(service string, realm BasicAuthRealm) (*client.Config, error) {
	res, err := initClient(service)
	if err != nil {
		return nil, err
	}
	conf := mustRealmConfig(realm)
	username := viper.GetString(fmt.Sprintf("%s-username", conf.ConfigKey))
	password := viper.GetString(fmt.Sprintf("%s-password", conf.ConfigKey))
	if username == "" || password == "" {
		return nil, errNoCredentials
	}
	allowInsecure := viper.GetBool("insecure")
	res.Credentials = client.BasicAuth(username, password, allowInsecure)
	return res, nil
}

// OAuth2Client returns a client configured with OAuth Client Credentials authentication.
func OAuth2Client(ctx context.Context, service string, scopes ...string) (*client.Config, error) {
	res, err := initClient(service)
	if err != nil {
		return nil, err
	}
	tokenURL := viper.GetString("token-url")
	clientID := viper.GetString("client-id")
	clientSecret := viper.GetString("client-secret")
	if clientID == "" || clientSecret == "" {
		return nil, errNoCredentials
	}
	audience := res.Address
	if host, _, err := net.SplitHostPort(audience); err == nil {
		audience = host
	}
	allowInsecure := viper.GetBool("insecure")
	res.Credentials = client.OAuth2(ctx, tokenURL, clientID, clientSecret, audience, scopes, allowInsecure)
	return res, nil
}

// ReadInConfig reads the configuration file with viper.
// A configuration file that cannot be found in the configured paths is not an error.
func ReadInConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}
	return nil
}

// AutomaticClient returns a client configured based on available settings.
// Basic authentication is preferred with the given realm.
// If Basic authentication is not configured, OAuth 2.0 Client Credentials are used.
func AutomaticClient(
	ctx context.Context, service string, basicAuthRealm BasicAuthRealm, oauthScopes ...string,
) (*client.Config, error) {
	for _, initFn := range []func() (*client.Config, error){
		func() (*client.Config, error) {
			return BasicAuthClient(service, basicAuthRealm)
		},
		func() (*client.Config, error) {
			return OAuth2Client(ctx, service, oauthScopes...)
		},
	} {
		config, err := initFn()
		if err == nil {
			return config, nil
		}
		if !errors.Is(err, errNoCredentials) {
			return nil, err
		}
	}
	return initClient(service)
}
