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

package messages_test

import (
	"bytes"
	"testing"

	"github.com/packetbroker/pb/pkg/messages"
)

func TestNetIDIssues(t *testing.T) {
	{
		netID := messages.NetID{0x00, 0x00, 0x13}

		good := messages.DevAddr{0x26, 0xff, 0x00, 0x00}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does issue %v", netID, good)
		}

		bad := messages.DevAddr{0x10, 0xff, 0x00, 0x00}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does not issue %v", netID, bad)
		}
	}

	{
		netID := messages.NetID{0x60, 0x00, 0x01}
		t.Log(netID.Type())

		good := messages.DevAddr{0xe0, 0x02, 0x00, 0x00}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does issue %v", netID, good)
		}

		good = messages.DevAddr{0xe0, 0x03, 0xff, 0xff}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does issue %v", netID, good)
		}

		bad := messages.DevAddr{0xe1, 0x02, 0x00, 0x00}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does not issue %v", netID, bad)
		}
	}

	{
		netID := messages.NetID{0xc0, 0x00, 0x01}
		t.Log(netID.Type())

		good := messages.DevAddr{0xfc, 0x00, 0x04, 0x00}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does issue %v", netID, good)
		}

		good = messages.DevAddr{0xfc, 0x00, 0x07, 0xff}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does issue %v", netID, good)
		}

		bad := messages.DevAddr{0xfc, 0x01, 0x04, 0x00}
		if !netID.Issues(good) {
			t.Errorf("NetID %v does not issue %v", netID, bad)
		}
	}
}

func TestBytes(t *testing.T) {
	b1 := messages.Bytes{0x00, 0xAA, 0xFF, 0x42}

	buf, err := b1.MarshalText()
	if err != nil {
		t.Fatalf("Failed to marshal (%v)", err)
	}
	if string(buf) != "00AAFF42" {
		t.Fatalf("Bytes do not marshal to `00AAFF42` but to `%s`", buf)
	}

	var b2 messages.Bytes
	if err := b2.UnmarshalText(buf); err != nil {
		t.Fatalf("Failed to unmarshal (%v)", err)
	}
	if !bytes.Equal(b1, b2) {
		t.Fatalf("The input `%v` does not equal unmarshaled `%v`", b1, b2)
	}
}
