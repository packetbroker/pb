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
	"encoding/json"
	"testing"
	"time"

	"github.com/packetbroker/pb/pkg/messages"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMarshaling(t *testing.T) {
	a := assertions.New(t)

	phyPayload := messages.PHYPayload{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}
	msg1 := messages.UplinkMessage{
		DevAddr:        messages.DevAddr{0x26, 0x01, 0x00, 0x51},
		PHYPayloadHash: messages.Bytes{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		PHYPayload:     &phyPayload,
		FCnt:           42,
		Time:           time.Date(2016, time.February, 12, 14, 24, 31, 0, time.UTC),
		DataRate:       "SF7BW125",
		Frequency:      868.1,
		FPort:          2,
	}

	buf, err := json.Marshal(msg1)
	a.So(err, should.BeNil)
	a.So(string(buf), should.Equal, `{"DevAddr":"26010051","FCntUp":42,"HPHYPayload":"0000000000000000000000000000000000000000000000000000000000000000","PHYPayload":"55555555555555555555","DataRate":"SF7BW125","ULFreq":868.1,"FPort":2,"RecvTime":1139322288}`)

	var msg2 messages.UplinkMessage
	err = json.Unmarshal(buf, &msg2)
	a.So(err, should.BeNil)
	a.So(msg1, should.Resemble, msg2)
}
