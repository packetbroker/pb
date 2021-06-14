// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	iampbv2 "go.packetbroker.org/api/iam/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
	"golang.org/x/term"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	networkAPIKeyCmd = &cobra.Command{
		Use:     "apikey",
		Aliases: []string{"apikeys", "key", "keys"},
		Short:   "Manage Packet Broker API keys for networks and tenants",
	}
	networkAPIKeyListCmd = &cobra.Command{
		Use:          "list",
		Aliases:      []string{"ls"},
		Short:        "List API keys",
		SilenceUsage: true,
		Example: `
  List all API keys:
    $ pbadmin network apikey list

  List API keys of a network:
    $ pbadmin network apikey list --net-id 000013

  List API keys of a tenant:
    $ pbadmin network apikey list --net-id 000013 --tenant-id tti

  List API keys of a named cluster in a network:
    $ pbadmin network apikey list --net-id 000013 --cluster-id eu1

  List API keys of a named cluster in a tenant:
    $ pbadmin network apikey list --net-id 000013 --tenant-id tti --cluster-id eu1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			hasNetID, hasTenantID, hasClusterID := pbflag.HasEndpoint(cmd.Flags(), "")
			endpoint := pbflag.GetEndpoint(cmd.Flags(), "")
			req := &iampbv2.ListNetworkAPIKeysRequest{}
			if hasNetID {
				req.NetId = wrapperspb.UInt32(uint32(endpoint.NetID))
			}
			if hasTenantID {
				req.TenantId = wrapperspb.String(endpoint.ID)
			}
			if hasClusterID {
				req.ClusterId = wrapperspb.String(endpoint.ClusterID)
			}
			res, err := iampbv2.NewNetworkAPIKeyVaultClient(conn).ListAPIKeys(ctx, req)
			if err != nil {
				return err
			}
			fmt.Fprintln(tabout, "Key ID\tNetID\tTenant ID\tCluster ID\tRights\tState\tLast Used\t")
			for _, t := range res.Keys {
				fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
					t.GetKeyId(),
					packetbroker.NetID(t.GetNetId()),
					t.GetTenantId(),
					t.GetClusterId(),
					column.Rights(t.GetRights()),
					t.GetState(),
					(*column.TimeSince)(t.GetAuthenticatedAt()),
				)
			}
			return nil
		},
	}
	networkAPIKeyCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an API key",
		Long: `Create an API key for a network or tenant, optionally for a named cluster.

The secret API key is returned only when creating the API key. You should store
the API key in a secure place, as it cannot be retrieved after create.`,
		Example: `
  Create API keys for a network:
    $ pbadmin network apikey create --net-id 000013

  Create API keys for a tenant:
    $ pbadmin network apikey create --net-id 000013 --tenant-id tti

  Create API keys for a named cluster in a network:
    $ pbadmin network apikey create --net-id 000013 --cluster-id eu1

  Create API key for a named cluster in a tenant:
    $ pbadmin network apikey create --net-id 000013 --tenant-id tti --cluster-id eu1

  Create API key for a network with rights to read networks and tenants:
    $ pbadmin network apikey create --net-id 000013 --rights READ_NETWORK

Rights:
  READ_NETWORK          Read networks
  READ_NETWORK_CONTACT  Read network contact information
  READ_TENANT           Read tenants
  READ_TENANT_CONTACT   Read tenant contact information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := pbflag.GetEndpoint(cmd.Flags(), "")
			req := &iampbv2.CreateNetworkAPIKeyRequest{
				NetId:     uint32(endpoint.NetID),
				TenantId:  endpoint.TenantID.ID,
				ClusterId: endpoint.ClusterID,
				Rights:    pbflag.GetAPIKeyRights(cmd.Flags()),
			}
			if promptKey, _ := cmd.Flags().GetBool("prompt-key"); promptKey {
				fmt.Fprint(os.Stdout, "Secret key: ")
				keyBuf, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return err
				}
				req.Key = string(keyBuf)
			}
			res, err := iampbv2.NewNetworkAPIKeyVaultClient(conn).CreateAPIKey(ctx, req)
			if err != nil {
				return err
			}
			if save, _ := cmd.Flags().GetBool("save"); save {
				viper.Set("client-id", res.Key.GetKeyId())
				viper.Set("client-secret", res.Key.GetKey())
				if err := viper.SafeWriteConfig(); err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, "Saved to API key to .pb.yaml")
			} else {
				fmt.Fprintln(os.Stderr, "Store the API key now in a secure place, as it cannot be retrieved later.")
			}
			return column.WriteKV(tabout,
				"Key ID", res.Key.GetKeyId(),
				"Secret Key", res.Key.GetKey(),
				"NetID", packetbroker.NetID(res.Key.GetNetId()),
				"Tenant ID", res.Key.GetTenantId(),
				"Cluster ID", res.Key.GetClusterId(),
				"Rights", column.Rights(res.Key.GetRights()),
				"State", res.Key.GetState(),
			)
		},
	}
	networkAPIKeyUpdateStateCmd = &cobra.Command{
		Use:          "update-state",
		Short:        "Update the API key state",
		SilenceUsage: true,
		Example: `
  Update the API key state to APPROVED:
    $ pbadmin network apikey update-state --key-id C5232IFFX4UKEELB --state APPROVED`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID, _ := cmd.Flags().GetString("key-id")
			state := pbflag.GetAPIKeyState(cmd.Flags(), "state")
			_, err := iampbv2.NewNetworkAPIKeyVaultClient(conn).UpdateAPIKeyState(ctx, &iampbv2.UpdateAPIKeyStateRequest{
				KeyId: keyID,
				State: state,
			})
			return err
		},
	}
	networkAPIKeyDeleteCmd = &cobra.Command{
		Use:          "delete",
		Aliases:      []string{"rm"},
		Short:        "Delete an API key",
		SilenceUsage: true,
		Example: `
  Delete an API key:
    $ pbadmin network apikey delete --key-id C5232IFFX4UKEELB`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID, _ := cmd.Flags().GetString("key-id")
			_, err := iampbv2.NewNetworkAPIKeyVaultClient(conn).DeleteAPIKey(ctx, &iampbv2.APIKeyRequest{
				KeyId: keyID,
			})
			return err
		},
	}
)

func init() {
	networkCmd.AddCommand(networkAPIKeyCmd)

	networkAPIKeyListCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	networkAPIKeyCmd.AddCommand(networkAPIKeyListCmd)

	networkAPIKeyCreateCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	networkAPIKeyCreateCmd.Flags().AddFlagSet(pbflag.APIKeyRights())
	networkAPIKeyCreateCmd.Flags().Bool("prompt-key", false, "prompt custom secret key value")
	networkAPIKeyCreateCmd.Flags().Bool("save", false, "save the API key to configuration")
	networkAPIKeyCmd.AddCommand(networkAPIKeyCreateCmd)

	networkAPIKeyUpdateStateCmd.Flags().String("key-id", "", "API key ID")
	networkAPIKeyUpdateStateCmd.Flags().AddFlagSet(pbflag.APIKeyState("state"))
	networkAPIKeyCmd.AddCommand(networkAPIKeyUpdateStateCmd)

	networkAPIKeyDeleteCmd.Flags().String("key-id", "", "API key ID")
	networkAPIKeyCmd.AddCommand(networkAPIKeyDeleteCmd)
}
