// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	iampb "go.packetbroker.org/api/iam"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	networkTenantCmd = &cobra.Command{
		Use:     "tenant",
		Aliases: []string{"tenants", "tnt", "tnts", "t"},
		Short:   "Manage Packet Broker tenants",
	}
	networkTenantListCmd = &cobra.Command{
		Use:          "list",
		Aliases:      []string{"ls"},
		Short:        "List tenants",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID := pbflag.GetNetID(cmd.Flags(), "")
			offset := uint32(0)
			fmt.Fprintln(tabout, "NetID\tTenant ID\tName\tDevAddr Blocks\t")
			for {
				res, err := iampb.NewTenantRegistryClient(conn).ListTenants(ctx, &iampb.ListTenantsRequest{
					NetId:  uint32(netID),
					Offset: offset,
				})
				if err != nil {
					return err
				}
				for _, t := range res.Tenants {
					fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t\n",
						packetbroker.NetID(t.GetNetId()),
						t.GetTenantId(),
						t.GetName(),
						column.DevAddrBlocks(t.GetDevAddrBlocks()),
					)
				}
				offset += uint32(len(res.Tenants))
				if len(res.Tenants) == 0 || offset >= res.Total {
					break
				}
			}
			return nil
		},
	}
	networkTenantCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a tenant",
		Example: `
  Create:
    $ pbadmin network tenant create --net-id 000013 --tenant-id tti

  Create with name:
    $ pbadmin network tenant create --net-id 000013 --tenant-id tti \
      --name "The Things Industries"

  Define DevAddr blocks to named clusters:
    $ pbadmin network tenant create --net-id 000013 --tenant-id tti \
      --dev-addr-blocks 26011000/20=eu1,26012000=eu2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID := pbflag.GetTenantID(cmd.Flags(), "")
			name, _ := cmd.Flags().GetString("name")
			devAddrBlocks := pbflag.GetDevAddrBlocks(cmd.Flags())
			res, err := iampb.NewTenantRegistryClient(conn).CreateTenant(ctx, &iampb.CreateTenantRequest{
				Tenant: &packetbroker.Tenant{
					NetId:         uint32(tenantID.NetID),
					TenantId:      tenantID.ID,
					Name:          name,
					DevAddrBlocks: devAddrBlocks,
					// TODO: Contact info (https://github.com/packetbroker/pb/issues/5)
				},
			})
			if err != nil {
				return err
			}
			return column.WriteTenant(tabout, res.Tenant)
		},
	}
	networkTenantGetCmd = &cobra.Command{
		Use:          "get",
		Short:        "Get a tenant",
		SilenceUsage: true,
		Example: `
  Get:
    $ pbadmin network tenant get --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID := pbflag.GetTenantID(cmd.Flags(), "")
			res, err := iampb.NewTenantRegistryClient(conn).GetTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			if err != nil {
				return err
			}
			return column.WriteTenant(tabout, res.Tenant)
		},
	}
	networkTenantUpdateCmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"up"},
		Short:   "Update a tenant",
		Example: `
  Update name:
    $ pbadmin network tenant update --net-id 000013 --tenant-id tti \
      --name "The Things Network"

  Define DevAddr blocks to named clusters:
    $ pbadmin network tenant update --net-id 000013 --tenant-id tti \
      --dev-addr-blocks 26011000/20=eu1,26012000=eu2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID := pbflag.GetTenantID(cmd.Flags(), "")
			req := &iampb.UpdateTenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
				// TODO: Contact info (https://github.com/packetbroker/pb/issues/5)
			}
			if cmd.Flags().Lookup("name").Changed {
				name, _ := cmd.Flags().GetString("name")
				req.Name = wrapperspb.String(name)
			}
			if cmd.Flags().Lookup("dev-addr-blocks").Changed {
				devAddrBlocks := pbflag.GetDevAddrBlocks(cmd.Flags())
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: devAddrBlocks,
				}
			}
			_, err := iampb.NewTenantRegistryClient(conn).UpdateTenant(ctx, req)
			return err
		},
	}
	networkTenantDeleteCmd = &cobra.Command{
		Use:          "delete",
		Aliases:      []string{"rm"},
		Short:        "Delete a tenant",
		SilenceUsage: true,
		Example: `
  Delete:
    $ pbadmin network tenant delete --net-id 000013 --tenant-id tti`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantID := pbflag.GetTenantID(cmd.Flags(), "")
			_, err := iampb.NewTenantRegistryClient(conn).DeleteTenant(ctx, &iampb.TenantRequest{
				NetId:    uint32(tenantID.NetID),
				TenantId: tenantID.ID,
			})
			return err
		},
	}
)

func tenantSettingsFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String("name", "", "tenant name")
	flags.AddFlagSet(pbflag.DevAddrBlocks())
	return flags
}

func init() {
	networkCmd.AddCommand(networkTenantCmd)

	networkTenantListCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkTenantCmd.AddCommand(networkTenantListCmd)

	networkTenantCreateCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantCreateCmd.Flags().AddFlagSet(tenantSettingsFlags())
	networkTenantCmd.AddCommand(networkTenantCreateCmd)

	networkTenantGetCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantCmd.AddCommand(networkTenantGetCmd)

	networkTenantUpdateCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantUpdateCmd.Flags().AddFlagSet(tenantSettingsFlags())
	networkTenantCmd.AddCommand(networkTenantUpdateCmd)

	networkTenantDeleteCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	networkTenantCmd.AddCommand(networkTenantDeleteCmd)
}
