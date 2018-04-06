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

package pbadmin

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/packetbroker/pb/pkg/messages"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var networkAddCmd = &cobra.Command{
	Use:   "add [net-id]",
	Short: "Add a network",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("pbadmin: subscribe needs exactly one argument")
			cmd.Help()
			os.Exit(1)
		}
		netID, err := messages.ParseNetID(args[0])
		if err != nil {
			fmt.Printf("pbadmin: invalid NetID (%v) (%#v)\n", err, args)
			os.Exit(1)
		}

		ctx := context.Background()
		client, err := pubsub.NewClient(ctx, viper.GetString("project-id"))
		if err != nil {
			fmt.Printf("pbadmin: failed to create client (%v)\n", err)
			os.Exit(1)
		}
		defer client.Close()

		createTopic := func(suffix string) *pubsub.Topic {
			id := fmt.Sprintf("netid-%v.%v", netID, suffix)
			topic, err := client.CreateTopic(ctx, id)
			if err != nil {
				if grpc.Code(err) == codes.AlreadyExists {
					fmt.Printf("pbadmin: topic %v already exists\n", id)
					return client.Topic(id)
				}
				fmt.Printf("pbadmin: failed to create topic (%v)\n", err)
				os.Exit(1)
			}
			return topic
		}
		inUp := createTopic("in.up")
		createTopic("in.keys")
		createTopic("out.up")

		createSubscription := func(prefix string, topic *pubsub.Topic, retention time.Duration) *pubsub.Subscription {
			id := fmt.Sprintf("%v-netid-%v", prefix, netID)
			sub, err := client.CreateSubscription(ctx, id, pubsub.SubscriptionConfig{
				Topic:             topic,
				RetentionDuration: retention,
			})
			if err != nil {
				if grpc.Code(err) == codes.AlreadyExists {
					fmt.Printf("pbadmin: subscription %v already exists\n", id)
					return client.Subscription(id)
				}
				fmt.Printf("pbadmin: failed to create subscription (%v)\n", err)
				os.Exit(1)
			}
			return sub
		}
		createSubscription("route", inUp, 0)
	},
}

func init() {
	networkCmd.AddCommand(networkAddCmd)
	viper.BindPFlags(networkAddCmd.PersistentFlags())
}
