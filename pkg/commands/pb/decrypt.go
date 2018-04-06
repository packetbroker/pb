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

package pb

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/packetbroker/pb/pkg/messages"
	"github.com/spf13/cobra"
)

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt a message received from Packet Broker",
	Run: func(cmd *cobra.Command, args []string) {
		hexKey, err := cmd.Flags().GetString("key")
		if err != nil {
			fmt.Printf("pb: failed to get key flag (%v)\n", err)
			os.Exit(1)
		}
		key, err := hex.DecodeString(hexKey)
		if err != nil {
			fmt.Printf("pb: failed to parse key (%v)\n", err)
			os.Exit(1)
		}
		if len(key) != 16 {
			fmt.Println("pb: key must be of length 16")
			os.Exit(1)
		}

		reader := bufio.NewReader(os.Stdin)
		for {
			s, err := reader.ReadString('\n')
			if err == io.EOF {
				return
			} else if err != nil {
				fmt.Printf("pb: failed to read input (%v)\n", err)
				os.Exit(1)
			}

			var uplink messages.UplinkMessage
			err = json.Unmarshal([]byte(s), &uplink)
			if err != nil {
				fmt.Printf("pb: failed to unmarshal JSON (%v)\n", err)
				os.Exit(1)
			}

			phyPayload, err := uplink.PHYPayloadEncrypted.VerifyAndDecrypt(key)
			if err != nil {
				fmt.Printf("pb: failed to decrypt PHYPayload (%v)\n", err)
				os.Exit(1)
			}
			uplink.PHYPayload = phyPayload
			uplink.PHYPayloadEncrypted = nil

			gatewayMetadata, err := uplink.GatewayMetadataEncrypted.VerifyAndDecrypt(key)
			if err != nil {
				fmt.Printf("pb: failed to decrypt gateway metadata (%v)\n", err)
				os.Exit(1)
			}
			uplink.GatewayMetadata = gatewayMetadata
			uplink.GatewayMetadataEncrypted = nil

			var buf []byte
			if val, _ := cmd.Flags().GetBool("pretty"); val {
				buf, err = json.MarshalIndent(uplink, "", "\t")
			} else {
				buf, err = json.Marshal(uplink)
			}
			if err != nil {
				fmt.Printf("pb: failed to marshal JSON (%v)\n", err)
				os.Exit(1)
			}

			fmt.Println(string(buf))
		}
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().String("key", "", "encryption key (hex)")
	decryptCmd.Flags().Bool("pretty", false, "show indented output")
}
