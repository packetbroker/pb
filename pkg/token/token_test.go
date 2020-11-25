// Copyright © 2020 The Things Industries B.V.

package token

import (
	"crypto/ed25519"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type keyProvider struct {
	keyID   string
	public  ed25519.PublicKey
	private ed25519.PrivateKey
}

func (k *keyProvider) PublicKeys() (*jose.JSONWebKeySet, error) {
	return &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Algorithm: "EdDSA",
				Key:       k.public,
				KeyID:     k.keyID,
			},
		},
	}, nil
}

func (k *keyProvider) PrivateKey() (*jose.JSONWebKey, error) {
	return &jose.JSONWebKey{
		Algorithm: "EdDSA",
		Key:       k.private,
		KeyID:     k.keyID,
	}, nil
}

func TestRoundtrip(t *testing.T) {
	Convey("Given a set of claims", t, func() {
		now := time.Now()
		claims := Claims{
			Claims: jwt.Claims{
				Issuer:   "test",
				IssuedAt: jwt.NewNumericDate(now),
				Expiry:   jwt.NewNumericDate(now.Add(1 * time.Hour)),
				Subject:  "test",
			},
			PacketBroker: &PacketBrokerClaims{
				Networks: []Network{
					{
						NetID:     0x000013,
						TenantID:  "tenant-a",
						ClusterID: "test",
					},
				},
			},
		}

		Convey("When generating a EdDSA key pair", func() {
			public, private, err := ed25519.GenerateKey(nil)
			So(err, ShouldBeNil)

			keyProvider := &keyProvider{"test", public, private}

			Convey("When signing the claims", func() {
				token, err := Sign(keyProvider, claims)
				So(err, ShouldBeNil)
				t.Logf("Token is %q", token)

				Convey("The parsed result should equal the claims", func() {
					parsed, err := Parse(keyProvider, token)
					So(err, ShouldBeNil)
					So(parsed, ShouldResemble, claims)
				})
			})
		})
	})
}
