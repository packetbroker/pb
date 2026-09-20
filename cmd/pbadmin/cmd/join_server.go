// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	iampb "github.com/packetbroker/go-api/iam/v2"
	packetbroker "github.com/packetbroker/go-api/v3"
	"github.com/packetbroker/pb/cmd/internal/column"
	"github.com/packetbroker/pb/cmd/internal/pbflag"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	joinServerCmd = &cobra.Command{
		Use:               "join-server",
		Aliases:           []string{"join-servers", "js"},
		Short:             "Manage Packet Broker Join Servers",
		PersistentPreRunE: prerunConnect,
		PersistentPostRun: postrunConnect,
	}
	joinServerListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Join Servers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				offset          = uint32(0)
				nameContains, _ = cmd.Flags().GetString("name-contains")
			)
			tabout.Println("  ID\tName\tJoinEUI Prefixes\tListed\tResolver\t")
			for {
				res, err := iampb.NewJoinServerRegistryClient(conn).ListJoinServers(ctx, &iampb.ListJoinServersRequest{
					Offset:       offset,
					NameContains: nameContains,
				})
				if err != nil {
					return fmt.Errorf("list Join Servers: %w", err)
				}
				for _, t := range res.GetJoinServers() {
					var resolver string
					if lookup := t.GetLookup(); lookup != nil {
						resolver = (*column.Target)(lookup).String()
					} else if fixed := t.GetFixed(); fixed != nil {
						resolver = (*column.JoinServerFixedEndpoint)(fixed).String()
					}
					tabout.Printf("%4d\t%s\t%s\t%s\t%s\t\n",
						t.GetId(),
						t.GetName(),
						column.JoinEUIPrefixes(t.GetJoinEuiPrefixes()),
						column.YesNo(t.GetListed()),
						resolver,
					)
				}
				offset += uint32(len(res.GetJoinServers()))
				if len(res.GetJoinServers()) == 0 || offset >= res.GetTotal() {
					break
				}
			}
			return nil
		},
	}
	joinServerCreateCmd = &cobra.Command{ //nolint:gosec // the example URL contains a placeholder password
		Use:   "create",
		Short: "Create a Join Server",
		Example: `
  Create:
    $ pbadmin join-server create --join-eui-prefixes EC656E0000000000/24

  Create with name and listed in the catalog:
    $ pbadmin join-server create --join-eui-prefixes EC656E0000000000/24 \
      --name "The Things Join Server" --listed

  Configure a LoRaWAN Backend Interfaces 1.1 target with HTTP basic auth:
    $ pbadmin join-server create --join-eui-prefixes EC656E0000000000/24 \
      --lookup-protocol TS002_V1_1 \
      --lookup-address https://user:pass@example.com

  See for more target configuration options:
    $ pbadmin join-server update target --help`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			name, _ := cmd.Flags().GetString("name")
			joinEUIPrefixes := pbflag.GetJoinEUIPrefixes(cmd.Flags())
			adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin")
			techContact := pbflag.GetContactInfo(cmd.Flags(), "tech")
			listed, _ := cmd.Flags().GetBool("listed")
			joinServer := &packetbroker.JoinServer{
				Name:                  name,
				JoinEuiPrefixes:       joinEUIPrefixes,
				AdministrativeContact: adminContact,
				TechnicalContact:      techContact,
				Listed:                listed,
			}
			if pbflag.TargetFlagsChanged(cmd.Flags(), "lookup") {
				var target *packetbroker.Target
				if err := pbflag.ApplyToTarget(cmd.Flags(), "lookup", &target); err != nil {
					return fmt.Errorf("configure target: %w", err)
				}
				joinServer.Resolver = &packetbroker.JoinServer_Lookup{
					Lookup: target,
				}
			} else if pbflag.EndpointFlagsChanged(cmd.Flags(), "fixed") {
				endpoint, _ := pbflag.GetEndpoint(cmd.Flags(), "fixed")
				joinServer.Resolver = &packetbroker.JoinServer_Fixed{
					Fixed: &packetbroker.JoinServerFixedEndpoint{
						NetId:     uint32(endpoint.NetID),
						TenantId:  endpoint.ID,
						ClusterId: endpoint.ClusterID,
					},
				}
			}
			res, err := iampb.NewJoinServerRegistryClient(conn).CreateJoinServer(ctx, &iampb.CreateJoinServerRequest{
				JoinServer: joinServer,
			})
			if err != nil {
				return fmt.Errorf("create Join Server: %w", err)
			}
			if err := column.WriteJoinServer(tabout, res.GetJoinServer(), false); err != nil {
				return fmt.Errorf("write Join Server: %w", err)
			}
			return nil
		},
	}
	joinServerGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a Join Server",
		Example: `
  Get:
    $ pbadmin join-server get --id 1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			joinServerID, _ := cmd.Flags().GetUint32("id")
			res, err := iampb.NewJoinServerRegistryClient(conn).GetJoinServer(ctx, &iampb.JoinServerRequest{
				Id: joinServerID,
			})
			if err != nil {
				return fmt.Errorf("get Join Server: %w", err)
			}
			verbose, _ := cmd.Flags().GetBool("verbose")
			if err := column.WriteJoinServer(tabout, res.GetJoinServer(), verbose); err != nil {
				return fmt.Errorf("write Join Server: %w", err)
			}
			return nil
		},
	}
	joinServerUpdateCmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"up"},
		Short:   "Update a Join Server",
		Example: `
  Update name:
    $ pbadmin join-server update --id 1 --name "The Things Join Server"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			joinServerID, _ := cmd.Flags().GetUint32("id")
			req := &iampb.UpdateJoinServerRequest{
				Id: joinServerID,
			}
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = wrapperspb.String(name)
			}
			if cmd.Flags().Changed("join-eui-prefixes") {
				joinEUIPrefixes := pbflag.GetJoinEUIPrefixes(cmd.Flags())
				req.JoinEuiPrefixes = &iampb.JoinEUIPrefixesValue{
					Value: joinEUIPrefixes,
				}
			}
			if adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin"); adminContact != nil {
				req.AdministrativeContact = &packetbroker.ContactInfoValue{
					Value: adminContact,
				}
			}
			if techContact := pbflag.GetContactInfo(cmd.Flags(), "tech"); techContact != nil {
				req.TechnicalContact = &packetbroker.ContactInfoValue{
					Value: techContact,
				}
			}
			if cmd.Flags().Changed("listed") {
				listed, _ := cmd.Flags().GetBool("listed")
				req.Listed = wrapperspb.Bool(listed)
			}
			_, err := iampb.NewJoinServerRegistryClient(conn).UpdateJoinServer(ctx, req)
			if err != nil {
				return fmt.Errorf("update Join Server: %w", err)
			}
			return nil
		},
	}
	joinServerUpdateTargetCmd = &cobra.Command{ //nolint:gosec // the example URL contains a placeholder password
		Use:   "target",
		Short: "Update a Join Server target",
		Example: `
  Configure a LoRaWAN Backend Interfaces 1.0 target with Packet Broker token
  authentication:
    $ pbadmin join-server update target --id 1 \
      --protocol TS002_V1_0 --address https://example.com --pb-token

  Configure a LoRaWAN Backend Interfaces 1.0 target with HTTP basic auth:
    $ pbadmin join-server update target --id 1 \
      --protocol TS002_V1_0 --address https://user:pass@example.com

  Configure a LoRaWAN Backend Interfaces 1.0 target with TLS:
    $ pbadmin join-server update target --id 1 \
      --protocol TS002_V1_0 --address https://example.com \
      --root-cas-file ca.pem --tls-cert-file key.pem --tls-key-file key.pem

  Configure a LoRaWAN Backend Interfaces 1.0 target with TLS and custom
  originating NetID:
    $ pbadmin join-server update target --id 1 --origin-net-id 000013 \
      --root-cas-file ca.pem --tls-cert-file key.pem --tls-key-file key.pem`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			joinServerID, _ := cmd.Flags().GetUint32("id")
			client := iampb.NewJoinServerRegistryClient(conn)
			joinServer, err := client.GetJoinServer(ctx, &iampb.JoinServerRequest{
				Id: joinServerID,
			})
			if err != nil {
				return fmt.Errorf("get Join Server: %w", err)
			}
			req := &iampb.UpdateJoinServerRequest{
				Id: joinServerID,
			}
			if pbflag.TargetFlagsChanged(cmd.Flags(), "lookup") {
				target := joinServer.GetJoinServer().GetLookup()
				if err := pbflag.ApplyToTarget(cmd.Flags(), "lookup", &target); err != nil {
					return fmt.Errorf("configure target: %w", err)
				}
				req.Resolver = &iampb.UpdateJoinServerRequest_Lookup{
					Lookup: target,
				}
			} else if pbflag.EndpointFlagsChanged(cmd.Flags(), "fixed") {
				endpoint, _ := pbflag.GetEndpoint(cmd.Flags(), "fixed")
				req.Resolver = &iampb.UpdateJoinServerRequest_Fixed{
					Fixed: &packetbroker.JoinServerFixedEndpoint{
						NetId:     uint32(endpoint.NetID),
						TenantId:  endpoint.ID,
						ClusterId: endpoint.ClusterID,
					},
				}
			}
			_, err = client.UpdateJoinServer(ctx, req)
			if err != nil {
				return fmt.Errorf("update Join Server: %w", err)
			}
			return nil
		},
	}
	joinServerDeleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "Delete a Join Server",
		Example: `
  Delete:
    $ pbadmin join-server delete --id 1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			joinServerID, _ := cmd.Flags().GetUint32("id")
			_, err := iampb.NewJoinServerRegistryClient(conn).DeleteJoinServer(ctx, &iampb.JoinServerRequest{
				Id: joinServerID,
			})
			if err != nil {
				return fmt.Errorf("delete Join Server: %w", err)
			}
			return nil
		},
	}
)

func joinServerSettingsFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String("name", "", "Join Server name")
	flags.Bool("listed", false, "list Join Server in catalog")
	return flags
}

func init() {
	rootCmd.AddCommand(joinServerCmd)

	joinServerListCmd.Flags().String("name-contains", "", "filter Join Servers by name")
	joinServerCmd.AddCommand(joinServerListCmd)

	joinServerCreateCmd.Flags().AddFlagSet(joinServerSettingsFlags())
	joinServerCreateCmd.Flags().AddFlagSet(pbflag.ContactInfo("admin"))
	joinServerCreateCmd.Flags().AddFlagSet(pbflag.ContactInfo("tech"))
	joinServerCreateCmd.Flags().AddFlagSet(pbflag.JoinEUIPrefixes())
	joinServerCreateCmd.Flags().AddFlagSet(pbflag.Target("lookup"))
	joinServerCreateCmd.Flags().AddFlagSet(pbflag.Endpoint("fixed"))
	joinServerCmd.AddCommand(joinServerCreateCmd)

	joinServerGetCmd.Flags().Uint32("id", 0, "unique identifier of the Join Server")
	joinServerGetCmd.Flags().Bool("verbose", false, "verbose output")
	joinServerCmd.AddCommand(joinServerGetCmd)

	joinServerUpdateCmd.Flags().Uint32("id", 0, "unique identifier of the Join Server")
	joinServerUpdateCmd.Flags().AddFlagSet(joinServerSettingsFlags())
	joinServerUpdateCmd.Flags().AddFlagSet(pbflag.JoinEUIPrefixes())
	joinServerUpdateCmd.Flags().AddFlagSet(pbflag.ContactInfo("admin"))
	joinServerUpdateCmd.Flags().AddFlagSet(pbflag.ContactInfo("tech"))
	joinServerUpdateTargetCmd.Flags().Uint32("id", 0, "unique identifier of the Join Server")
	joinServerUpdateTargetCmd.Flags().AddFlagSet(pbflag.Target("lookup"))
	joinServerUpdateTargetCmd.Flags().AddFlagSet(pbflag.Endpoint("fixed"))
	joinServerUpdateCmd.AddCommand(joinServerUpdateTargetCmd)
	joinServerCmd.AddCommand(joinServerUpdateCmd)

	joinServerDeleteCmd.Flags().Uint32("id", 0, "unique identifier of the Join Server")
	joinServerCmd.AddCommand(joinServerDeleteCmd)
}
