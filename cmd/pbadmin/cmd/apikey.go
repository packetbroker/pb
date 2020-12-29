// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	iampb "go.packetbroker.org/api/iam"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
)

var (
	apiKeyCmd = &cobra.Command{
		Use:     "apikey",
		Aliases: []string{"apikeys", "key", "keys"},
		Short:   "Manage Packet Broker API keys for networks and tenants",
	}
	apiKeyListCmd = &cobra.Command{
		Use:          "list",
		Aliases:      []string{"ls"},
		Short:        "List API keys",
		SilenceUsage: true,
		Example: `
  List API keys of a network:
    $ pbadmin apikey list --net-id 000013

  List API keys of a tenant:
    $ pbadmin apikey list --net-id 000013 --tenant-id tti

  List API keys of a named cluster in a network:
    $ pbadmin apikey list --net-id 000013 --cluster-id eu1

  List API keys of a named cluster in a tenant:
    $ pbadmin apikey list --net-id 000013 --tenant-id tti --cluster-id eu1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := pbflag.GetEndpoint(cmd.Flags(), "")
			res, err := iampb.NewAPIKeyVaultClient(conn).ListAPIKeys(ctx, &iampb.ListAPIKeysRequest{
				NetId:     uint32(endpoint.NetID),
				TenantId:  endpoint.TenantID.ID,
				ClusterId: endpoint.ClusterID,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(tabout, "Key ID\tNetID\tTenant ID\tCluster ID\tLast Used\t")
			for _, t := range res.Keys {
				fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t%s\t\n",
					t.GetKeyId(),
					packetbroker.NetID(t.GetNetId()),
					t.GetTenantId(),
					t.GetClusterId(),
					(*column.TimeSince)(t.GetAuthenticatedAt()),
				)
			}
			return nil
		},
	}
	apiKeyCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an API key",
		Long: `Create an API key for a network or tenant, optionally for a named cluster.

The secret API key is returned only when creating the API key. You should store
the API key in a secure place, as it cannot be retrieved after create.`,
		Example: `
  Create API keys for a network:
    $ pbadmin apikey create --net-id 000013

  Create API keys for a tenant:
    $ pbadmin apikey create --net-id 000013 --tenant-id tti

  Create API keys for a named cluster in a network:
    $ pbadmin apikey create --net-id 000013 --cluster-id eu1

  Create API keys for a named cluster in a tenant:
    $ pbadmin apikey create --net-id 000013 --tenant-id tti --cluster-id eu1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := pbflag.GetEndpoint(cmd.Flags(), "")
			res, err := iampb.NewAPIKeyVaultClient(conn).CreateAPIKey(ctx, &iampb.CreateAPIKeyRequest{
				NetId:     uint32(endpoint.NetID),
				TenantId:  endpoint.TenantID.ID,
				ClusterId: endpoint.ClusterID,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "Store the API key now in a secure place, as it cannot be retrieved later.")
			return column.WriteKV(tabout,
				"Key ID", res.Key.GetKeyId(),
				"Secret Key", res.Key.GetKey(),
				"NetID", packetbroker.NetID(res.Key.GetNetId()),
				"Tenant ID", res.Key.GetTenantId(),
				"Cluster ID", res.Key.GetClusterId(),
			)
		},
	}
	apiKeyDeleteCmd = &cobra.Command{
		Use:          "delete",
		Aliases:      []string{"rm"},
		Short:        "Delete an API key",
		SilenceUsage: true,
		Example: `
  Delete an API key:
    $ pbadmin apikey delete --key-id C5232IFFX4UKEELB`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID, _ := cmd.Flags().GetString("key-id")
			_, err := iampb.NewAPIKeyVaultClient(conn).DeleteAPIKey(ctx, &iampb.APIKeyRequest{
				KeyId: keyID,
			})
			return err
		},
	}
)

func init() {
	rootCmd.AddCommand(apiKeyCmd)

	apiKeyListCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	apiKeyCmd.AddCommand(apiKeyListCmd)

	apiKeyCreateCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	apiKeyCmd.AddCommand(apiKeyCreateCmd)

	apiKeyDeleteCmd.Flags().String("key-id", "", "API key ID")
	apiKeyCmd.AddCommand(apiKeyDeleteCmd)
}
