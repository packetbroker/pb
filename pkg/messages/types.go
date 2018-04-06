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

// package messages provides message formats and interfaces for the LoRaWAN Packet Broker.
package messages

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// DevAddr is the device's four byte address.
type DevAddr [4]byte

// ParseDevAddr parses the specified hex encoded string.
func ParseDevAddr(s string) (devAddr DevAddr, err error) {
	buf, err := hex.DecodeString(string(s))
	if err != nil {
		return
	}
	if len(buf) != 4 {
		err = fmt.Errorf("invalid DevAddr %v: expecting 4 bytes", s)
		return
	}
	copy(devAddr[:], buf)
	return
}

// MarshalText implements the TextMarshaler interface.
func (d DevAddr) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (d *DevAddr) UnmarshalText(s []byte) error {
	val, err := ParseDevAddr(string(s))
	if err != nil {
		return err
	}
	*d = val
	return nil
}

// String implements the Stringer interface.
func (d DevAddr) String() string {
	return strings.ToUpper(hex.EncodeToString(d[:]))
}

// NetID is the network's unique identifier, issued by the LoRa Alliance.
type NetID [3]byte

// ParseNetID parses the specified hex encoded string.
func ParseNetID(s string) (netID NetID, err error) {
	buf, err := hex.DecodeString(string(s))
	if err != nil {
		return
	}
	if len(buf) != 3 {
		err = fmt.Errorf("invalid NetID %v: expecting 3 bytes", s)
		return
	}
	copy(netID[:], buf)
	return
}

// MarshalText implements the TextMarshaler interface.
func (n NetID) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (n *NetID) UnmarshalText(s []byte) error {
	val, err := ParseNetID(string(s))
	if err != nil {
		return err
	}
	*n = val
	return nil
}

// String implements the Stringer interface.
func (n NetID) String() string {
	return strings.ToUpper(hex.EncodeToString(n[:]))
}

// Type returns the NetID type.
func (n NetID) Type() byte {
	return n[0] >> 5
}

// Dec returns the decimal value.
func (n NetID) Dec() int {
	return int(n[0])<<16 | (int(n[1]) << 8) | int(n[2])
}

// Issues returns whether the NetID issues the specified DevAddr.
func (n NetID) Issues(d DevAddr) bool {
	switch n.Type() {
	case 0:
		return n[2]&0x7f == d[0]>>1
	case 3:
		id := n.Dec() & 0x1fffff
		return d[0] == 0xe0|byte((id>>7)&0x0f) && d[1]&0xfe == byte(id&0xff)<<1
	case 6:
		id := n.Dec() & 0x1fffff
		return d[0] == 0xfc|byte((id>>14)&0x01) && d[1]&0xff == byte(id>>6) && d[2]&0xfc == byte(id&0x3f)<<2
	default:
		return false
	}
}

// EncryptionKey contains an encryption key and its identifier.
type EncryptionKey struct {
	ID    string
	Value []byte
}

// Bytes is a byte slice that gets text marshaled to hex.
type Bytes []byte

// MarshalText implements the TextMarshaler interface.
func (b Bytes) MarshalText() ([]byte, error) {
	return []byte(strings.ToUpper(hex.EncodeToString(b))), nil
}

// UnmarshalText implements the TextMarshaler interface.
func (b *Bytes) UnmarshalText(s []byte) error {
	buf, err := hex.DecodeString(string(s))
	if err != nil {
		return err
	}
	*b = Bytes(buf[:])
	return nil
}
