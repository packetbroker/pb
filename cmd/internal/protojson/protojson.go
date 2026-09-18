// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package protojson provides JSON encoding and decoding of proto messages with the default options for Packet Broker.
package protojson

import (
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var marshalOptions = protojson.MarshalOptions{
	Multiline:       true,
	Indent:          "  ",
	AllowPartial:    true,
	UseProtoNames:   false,
	UseEnumNumbers:  false,
	EmitUnpopulated: true,
}

// Marshal marshals the proto message using the default options for Packet Broker.
func Marshal(m proto.Message) ([]byte, error) {
	res, err := marshalOptions.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}
	return res, nil
}

// Write marshals the proto message (see Marshal) and writes it to the given writer.
func Write(w io.Writer, m proto.Message) error {
	rawMsg, err := Marshal(m)
	if err != nil {
		return err
	}
	if _, err := w.Write(rawMsg); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	return nil
}

var unmarshalOptions = protojson.UnmarshalOptions{
	AllowPartial: true,
}

// Unmarshal unmarshals the proto message using the default options for Packet Broker.
func Unmarshal(b []byte, m proto.Message) error {
	if err := unmarshalOptions.Unmarshal(b, m); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}

// Decode reads a JSON message from the JSON decoder and unmarshals it (see Unmarshal).
func Decode(d *json.Decoder, m proto.Message) error {
	var rawMsg json.RawMessage
	if err := d.Decode(&rawMsg); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	return Unmarshal(rawMsg, m)
}
