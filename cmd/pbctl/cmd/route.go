// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	routingpb "go.packetbroker.org/api/routing"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var routeCmd = &cobra.Command{Use: "route", Aliases: []string{"routes", "ro"}, Short: "List Packet Broker routes", RunE: func(cmd *cobra.Command, args []string) error {
	var (
		client        = routingpb.NewRoutesClient(conn)
		lastCreatedAt time.Time
	)
	for {
		res, err := client.ListRoutes(ctx, &routingpb.ListRoutesRequest{CreatedSince: timestamppb.New(lastCreatedAt)})
		if err != nil {
			return err
		}
		for _, p := range res.Routes {
			if err := protojson.Write(os.Stdout, p); err != nil {
				return err
			}
			lastCreatedAt = p.CreatedAt.AsTime()
		}
		if len(res.Routes) == 0 {
			break
		}
	}
	return nil
}}

func init() {
	rootCmd.AddCommand(routeCmd)
}
