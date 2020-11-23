// Copyright © 2020 The Things Industries B.V.

package protojson

import (
	"encoding/json"
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
	return marshalOptions.Marshal(m)
}

// Write marshals the proto message (see Marshal) and writes it to the given writer.
func Write(w io.Writer, m proto.Message) error {
	rawMsg, err := Marshal(m)
	if err != nil {
		return err
	}
	_, err = w.Write(rawMsg)
	return err
}

var unmarshalOptions = protojson.UnmarshalOptions{
	AllowPartial: true,
}

// Unmarshal unmarshals the proto message using the default options for Packet Broker.
func Unmarshal(b []byte, m proto.Message) error {
	return unmarshalOptions.Unmarshal(b, m)
}

// Decode reads a JSON message from the JSON decoder and unmarshals it (see Unmarshal).
func Decode(d *json.Decoder, m proto.Message) error {
	var rawMsg json.RawMessage
	if err := d.Decode(&rawMsg); err != nil {
		return err
	}
	return Unmarshal(rawMsg, m)
}
