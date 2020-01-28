// Copyright © 2020 The Things Industries B.V.

package console

import (
	"os"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

var marshaler = &jsonpb.Marshaler{
	EmitDefaults: true,
	Indent:       "  ",
}

// WriteProto writes the proto message to os.Stdout.
func WriteProto(res proto.Message) error {
	return marshaler.Marshal(os.Stdout, res)
}
