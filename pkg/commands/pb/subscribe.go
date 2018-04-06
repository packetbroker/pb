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
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/packetbroker/pb/pkg/messages"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// subscribeCmd represents the subscribe command
var subscribeCmd = &cobra.Command{
	Use:   "subscribe [net-id]",
	Short: "Subscribe to messages received from Packet Broker",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("pb: subscribe needs exactly one argument")
			cmd.Help()
			os.Exit(1)
		}
		netID, err := messages.ParseNetID(args[0])
		if err != nil {
			fmt.Printf("pb: invalid NetID (%v)\n", netID)
			os.Exit(1)
		}

		ctx := context.Background()
		client, err := pubsub.NewClient(ctx, viper.GetString("project-id"))
		if err != nil {
			fmt.Printf("pb: failed to create client (%v)\n", err)
			os.Exit(1)
		}

		subscriptionID := fmt.Sprintf("debug-netid-%v", netID)
		subscription := client.Subscription(subscriptionID)
		err = subscription.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
			var uplink messages.UplinkMessage
			if lerr := json.Unmarshal(message.Data, &uplink); lerr != nil {
				fmt.Printf("pb: failed to unmarshal JSON (%v)\n", lerr)
				message.Nack()
				return
			}

			var lerr error
			var buf []byte
			if val, _ := cmd.Flags().GetBool("pretty"); val {
				buf, lerr = json.MarshalIndent(uplink, "", "\t")
			} else {
				buf, lerr = json.Marshal(uplink)
			}
			if lerr != nil {
				fmt.Printf("pb: failed to marshal JSON (%v)\n", lerr)
				message.Nack()
				return
			}

			fmt.Println(string(buf))
			message.Ack()
		})
		if err != nil {
			fmt.Printf("pb: failed to receive from subscription %v (%v)\n", subscriptionID, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(subscribeCmd)

	subscribeCmd.Flags().Bool("pretty", false, "show indented output")
}
