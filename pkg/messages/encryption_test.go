// Copyright © 2018 The Packet Broker Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package messages

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPHYPayloadEncryption(t *testing.T) {
	a := assertions.New(t)

	key := EncryptionKey{
		ID:    "test",
		Value: make([]byte, 16),
	}
	io.ReadFull(rand.Reader, key.Value)

	payload1 := PHYPayload{0x01, 0x19, 0x04, 0x40}

	enc, err := payload1.EncryptAndSign(key)
	a.So(err, should.BeNil)
	a.So(enc.KeyID, should.Equal, "test")

	payload2, err := enc.VerifyAndDecrypt(key.Value)
	a.So(err, should.BeNil)
	a.So(*payload2, should.Resemble, payload1)
}

func TestGatewayMetadataEncryption(t *testing.T) {
	a := assertions.New(t)

	key := EncryptionKey{
		ID:    "test",
		Value: make([]byte, 16),
	}
	io.ReadFull(rand.Reader, key.Value)

	metadata1 := GatewayMetadata{
		ID: "test",
		Antennas: []AntennaMetadata{
			{
				RSSI: -156,
			},
		},
	}

	enc, err := metadata1.EncryptAndSign(key)
	a.So(err, should.BeNil)
	a.So(enc.KeyID, should.Equal, "test")

	metadata2, err := enc.VerifyAndDecrypt(key.Value)
	a.So(err, should.BeNil)
	a.So(*metadata2, should.Resemble, metadata1)
}
