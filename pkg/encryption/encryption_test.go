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

package encryption_test

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	"github.com/packetbroker/pb/pkg/encryption"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestEncryption(t *testing.T) {
	a := assertions.New(t)

	// Test length: 15 bytes has 1 byte for padding
	{
		secret := "123456789012345"
		fmt.Println(len(secret))
		key := make([]byte, 16)
		io.ReadFull(rand.Reader, key)

		buf := []byte(secret)
		encrypted, err := encryption.Encrypt(buf, key)
		fmt.Println("encrypted buf:", encrypted)
		a.So(err, should.BeNil)
		a.So(encrypted, should.HaveLength, aes.BlockSize+aes.BlockSize)
	}

	// Test length: 16 bytes need another block for padding
	{
		secret := "1234567890123456"
		key := make([]byte, 16)
		io.ReadFull(rand.Reader, key)

		buf := []byte(secret)
		encrypted, err := encryption.Encrypt(buf, key)
		fmt.Println("encrypted buf:", encrypted)
		a.So(err, should.BeNil)
		a.So(encrypted, should.HaveLength, aes.BlockSize+2*aes.BlockSize)
	}

	// Test roundtrip
	{
		secret := "1234567890123456789012345678901234567890"
		key := make([]byte, 16)
		io.ReadFull(rand.Reader, key)

		buf1 := []byte(secret)
		encrypted, err := encryption.Encrypt(buf1, key)
		fmt.Println("encrypted buf:", encrypted)
		a.So(err, should.BeNil)
		a.So(encrypted, should.HaveLength, aes.BlockSize+3*aes.BlockSize)

		buf2, err := encryption.Decrypt(encrypted, key)
		a.So(err, should.BeNil)
		a.So(buf2, should.Resemble, buf1)
	}
}
