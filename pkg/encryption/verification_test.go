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
	"crypto/rand"
	"io"
	"testing"

	"github.com/packetbroker/pb/pkg/encryption"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestVerification(t *testing.T) {
	a := assertions.New(t)

	data := "some encrypted data"
	key := make([]byte, 16)
	io.ReadFull(rand.Reader, key)

	mac, err := encryption.Sign([]byte(data), key)
	a.So(err, should.BeNil)

	ok, err := encryption.Verify([]byte(data), mac, key)
	a.So(err, should.BeNil)
	a.So(ok, should.BeTrue)

	ok, err = encryption.Verify([]byte{0x1, 0x2}, mac, key)
	a.So(err, should.BeNil)
	a.So(ok, should.BeFalse)
}
