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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/packetbroker/pb/pkg/encryption"
)

func encryptAndSign(buf []byte, key EncryptionKey) (*EncryptedData, error) {
	data, err := encryption.Encrypt(buf, key.Value)
	if err != nil {
		return nil, fmt.Errorf("encrypt failed: %v", err)
	}
	mac, err := encryption.Sign(data, key.Value)
	if err != nil {
		return nil, fmt.Errorf("sign failed: %v", err)
	}
	return &EncryptedData{
		Data:  data,
		MAC:   mac,
		KeyID: key.ID,
	}, nil
}

func (d EncryptedData) verifyAndDecrypt(key []byte) ([]byte, error) {
	if ok, err := encryption.Verify(d.Data, d.MAC, key); err != nil {
		return nil, fmt.Errorf("verify failed: %v", err)
	} else if !ok {
		return nil, errors.New("verify failed: invalid MAC")
	}
	buf, err := encryption.Decrypt(d.Data, key)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %v", err)
	}
	return buf, nil
}

// EncryptAndSign encrypts and signs the PHYPayload and calculates the SHA-256 hash.
func (p PHYPayload) EncryptAndSign(key EncryptionKey) (*PHYPayloadEncrypted, error) {
	data, err := encryptAndSign(p, key)
	if err != nil {
		return nil, err
	}
	return &PHYPayloadEncrypted{
		EncryptedData: *data,
	}, nil
}

// VerifyAndDecrypt verifies the MAC and decrypts the PHYPayload.
func (p PHYPayloadEncrypted) VerifyAndDecrypt(key []byte) (*PHYPayload, error) {
	buf, err := p.verifyAndDecrypt(key)
	if err != nil {
		return nil, err
	}
	payload := PHYPayload(buf)
	return &payload, nil
}

// EncryptAndSign encrypts and signs the GatewayMetadata.
func (p GatewayMetadata) EncryptAndSign(key EncryptionKey) (*GatewayMetadataEncrypted, error) {
	buf, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	data, err := encryptAndSign(buf, key)
	if err != nil {
		return nil, err
	}
	return &GatewayMetadataEncrypted{
		EncryptedData: *data,
	}, nil
}

// VerifyAndDecrypt verifies the MAC and decrypts the GatewayMetadata.
func (p GatewayMetadataEncrypted) VerifyAndDecrypt(key []byte) (*GatewayMetadata, error) {
	buf, err := p.verifyAndDecrypt(key)
	if err != nil {
		return nil, err
	}
	var metadata GatewayMetadata
	if err := json.Unmarshal(buf, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}
