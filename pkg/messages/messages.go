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
	"time"

	"github.com/packetbroker/pb/pkg/gpstime"
)

// EncryptedData contains encrypted data and the encryption key ID.
type EncryptedData struct {
	Data        Bytes  `json:"Data"`
	MAC         Bytes  `json:"MAC"`
	KeyID       string `json:"KeyID"`
	KeyExchange string `json:"KeyExchange"`
}

// PHYPayload is the physical LoRaWAN payload.
type PHYPayload Bytes

// MarshalText implements the TextMarshaler interface.
func (p PHYPayload) MarshalText() ([]byte, error) {
	return Bytes(p).MarshalText()
}

// UnmarshalText implements the TextMarshaler interface.
func (p *PHYPayload) UnmarshalText(s []byte) error {
	var b Bytes
	if err := b.UnmarshalText(s); err != nil {
		return err
	}
	*p = PHYPayload(b)
	return nil
}

// PHYPayloadEncrypted contains the hash and encrypted PHYPayload.
type PHYPayloadEncrypted struct {
	EncryptedData
}

// AntennaMetadata contains received metadata from a gateway antenna.
type AntennaMetadata struct {
	ID                     uint32   `json:"ID"`
	RSSI                   float32  `json:"RSSI"`
	SNR                    float32  `json:"SNR"`
	Latitude               *float64 `json:"Lat,omitempty"`
	Longitude              *float64 `json:"Lon,omitempty"`
	Altitude               *float64 `json:"Alt,omitempty"`
	FineTimestamp          *uint64  `json:"FTime,omitempty"`
	FineTimestampEncrypted Bytes    `json:"ETime,omitempty"`
}

// GatewayMetadata contains received metadata from a gateway.
type GatewayMetadata struct {
	ID              string            `json:"ID"`
	Region          string            `json:"RFRegion,omitempty"`
	UplinkToken     string            `json:"ULToken,omitempty"`
	DownlinkAllowed bool              `json:"DLAllowed"`
	Antennas        []AntennaMetadata `json:"Antennas"`
}

// GatewayMetadataEncrypted contains the encrypted gateway metadata.
type GatewayMetadataEncrypted struct {
	EncryptedData
}

// UplinkMessage is an uplink message received from a gateway.
type UplinkMessage struct {
	Origin                   *NetID                    `json:"Origin,omitempty"`
	DevAddr                  DevAddr                   `json:"DevAddr"`
	FCnt                     uint32                    `json:"FCntUp"`
	PHYPayloadHash           Bytes                     `json:"HPHYPayload"`
	PHYPayload               *PHYPayload               `json:"PHYPayload,omitempty"`
	PHYPayloadEncrypted      *PHYPayloadEncrypted      `json:"EPHYPayload,omitempty"`
	Time                     time.Time                 `json:"-"`
	DataRate                 string                    `json:"DataRate"`
	Frequency                float64                   `json:"ULFreq"`
	FPort                    uint8                     `json:"FPort"`
	GatewayMetadata          *GatewayMetadata          `json:"GWInfo,omitempty"`
	GatewayMetadataEncrypted *GatewayMetadataEncrypted `json:"EGWInfo,omitempty"`
}

// MarshalJSON implements the Marshaler interface.
func (m UplinkMessage) MarshalJSON() ([]byte, error) {
	type Aux UplinkMessage
	return json.Marshal(struct {
		Aux
		Time int64 `json:"RecvTime"`
	}{
		Aux:  Aux(m),
		Time: gpstime.ToGPS(m.Time),
	})
}

// UnmarshalJSON implements the Marshaler interface.
func (m *UplinkMessage) UnmarshalJSON(data []byte) error {
	type Aux UplinkMessage
	aux := struct {
		*Aux
		Time int64 `json:"RecvTime"`
	}{
		Aux: (*Aux)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	m.Time = gpstime.Parse(aux.Time).UTC()
	return nil
}
