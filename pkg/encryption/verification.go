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

package encryption

import (
	"crypto/hmac"
	"crypto/sha256"
)

// Sign returns the MAC of data with key using HMAC with SHA-256.
func Sign(data, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}

// Verify computes the MAC using Sign, and returns true when the MACs are equal.
func Verify(data, actual, key []byte) (bool, error) {
	expected, err := Sign(data, key)
	if err != nil {
		return false, err
	}
	return hmac.Equal(expected, actual), nil
}
