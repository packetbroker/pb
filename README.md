# Packet Broker Clients

Packet Broker Clients are command-line utilities for working with [Packet Broker](https://www.packetbroker.org).

## Services

Action | Service | Client | Basic Auth | OAuth 2.0
--- | --- | --- | --- | ---
Manage networks | IAM | `pbadmin` | administrator |
Manage tenants | IAM | `pbadmin` | administrator | network
Manage network API keys | IAM | `pbadmin` | administrator | network
Manage cluster API keys | IAM | `pbadmin` | administrator |
List networks and tenants | IAM | `pbadmin` | | cluster, network
Manage routing policies | Control Plane | `pbctl` | | network, tenant
List routes | Control Plane | `pbctl` | | cluster, network, tenant
List routing policies | Control Plane | `pbctl` | | cluster, network, tenant
Publish and subscribe | Data Plane | `pbpub`, `pbsub` | | network, tenant

IAM and Control Plane are deployed in a global cluster. Routers (with Data Plane) are deployed in regional clusters:

Region | Address
--- | ---
Europe | `eu.packetbroker.io:443`
North America | `nam.packetbroker.io:443`
Asia Pacific | `apac.packetbroker.io:443`

## Getting Started

### Installation

Make sure you have [Go](https://golang.org/doc/install) installed in your environment.

```bash
$ go get go.packetbroker.org/pb/cmd/pbadmin
$ go get go.packetbroker.org/pb/cmd/pbctl
$ go get go.packetbroker.org/pb/cmd/pbpub
$ go get go.packetbroker.org/pb/cmd/pbsub
```

### Configuration

Create a configuration file `$HOME/.pb.yaml` or `.pb.yaml` in the working directory:

```yaml
# Router region:
router-address: "eu.packetbroker.io:443"

# Network or tenant API key ID and secret key value:
client-id: "KZUCD5XAYT6EJ5BH"
client-secret: "E67X5675UCQFTTJMUD73URQOLPA5VT4GBFLPCMUHZWK52ML5"

# Uncomment if using pbadmin with full administrative access:
#iam-username: "admin"
#iam-password: "admin"
```

### Command-Line Interface

The command-line utilities `pbadmin`, `pbctl`, `pbpub` and `pbsub` contain extensive examples. Specify `--help` to show examples and possible flags.

### Manage Network Tenants

Packet Broker Identity and Access Management (IAM) stores networks and tenants. Networks are LoRaWAN networks with a NetID, i.e. `000013` (with DevAddr prefix `26000000/7`). Tenants make use of one or more DevAddr blocks within a NetID, i.e. NetID `000013` with prefix `26AA0000/16`. Tenants have a unique identifier within the NetID, called the tenant ID.

As NetID `000013`, to create tenant `tenant-a` with DevAddr blocks `26AA0000/16` and `26BB0000/16`:

```bash
$ pbadmin network tenant create --net-id 000013 --tenant-id tenant-a \
    --name "Tenant A" --dev-addr-blocks 26AA0000/16,26BB0000/16
```

>The prefixes indicate the base DevAddr and the bit length, like [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation). For example, `26AA0000/16` matches all DevAddr from `26AA0000` to `26AAFFFF`.

Optionally, you can specify a Home Network cluster. If your network uses multiple clusters, you can let Packet Broker route traffic to these clusters:

```bash
$ pbadmin network tenant update --net-id 000013 --tenant-id tenant a \
    --dev-addr-blocks 26AA0000/16=eu1,26BB0000/16=eu2
```

To list tenants:

```bash
$ pbadmin network tenant list --net-id 000013
```

To get a tenant:

```bash
$ pbadmin network tenant get --net-id 000013 --tenant-id tenant-a
```

To delete a tenant:

```bash
$ pbadmin network tenant delete --net-id 000013 --tenant-id tenant-a
```

### Manage Network and Tenant API Keys

Network API keys are used by clients of Packet Broker Router: LoRaWAN network servers and the command-line utilities `pbpub` and `pbsub`.

You can create API keys for a network, a tenant, a named cluster in a network and a named cluster in a tenant. For example, to create an API key for a tenant:

```bash
$ pbadmin network apikey create --net-id 000042 --tenant-id tenant-a
```

And for a named cluster in a tenant:

```bash
$ pbadmin network apikey create --net-id 000042 --tenant-id tenant-a --cluster-id eu1
```

### Configure Routing Policies

As a Forwarder, you can configure a default routing policy for all Home Networks and routing policies per Home Network with `pbctl`. 

As Forwarder NetID `000042`, to see the default routing policy:

```bash
$ pbctl policy get --forwarder-net-id 000042 --defaults
```

To see the routing policy between Forwarder `000042` and Home Network NetID `C00123`:

```bash
$ pbctl policy get --forwarder-net-id 000042 --home-network-net-id C00123
```

To see the routing policy between Forwarder NetID `000042` tenant `tenant-a` and Home Network NetID `C00123` tenant `tenant-b`:

```bash
$ pbctl policy get --forwarder-net-id 000042 --forwarder-tenant-id tenant-a \
    --home-network-net-id C00123 --home-network-tenant-id tenant-b
```

To see all the routing policies that Forwarders configured for Home Network NetID `C00123`:

```bash
$ pbctl policy list --home-network-net-id C00123
```

You can set policies by specifying letters from the following table:

| Policy | Uplink | Downlink |
| --- | :---: | :---: |
| Join-request/accept | `J` | `J` |
| MAC commands | `M` | `M` |
| Application data | `A` | `A` |
| Signal Quality | `S` | |
| Localization | `L` | |

To enable all exchange by default:

```bash
$ pbctl policy set --forwarder-net-id 000042 --defaults \
    --set-uplink JMASL --set-downlink --JMA
```

To enable only device activation and MAC commands in both directions with Home Network NetID `C00123`:

```bash
$ pbctl policy set --forwarder-net-id 000042 --home-network-net-id C00123 \
    --set-uplink JM --set-downlink --JM
```

To enable only device activation and MAC commands in both directions of Forwarder tenant `tenant-a` with Home Network NetID `C00123` tenant `tenant-b`:

```bash
$ pbctl policy set --forwarder-net-id 000042 --forwarder-tenant-id tenant-a \
    --home-network-net-id C00123 --home-network-tenant-id tenant-b \
    --set-uplink JM --set-downlink --JM
```

### Publish and Subscribe Traffic

To subscribe to routed downlink traffic as network, tenant, and with or without named cluster:

```bash
$ pbsub --forwarder-net-id 000042 --group debug
$ pbsub --forwarder-net-id 000042 --forwarder-tenant-id tenant-a --group debug
$ pbsub --forwarder-net-id 000042 --forwarder-cluster-id eu1 --group debug
$ pbsub --forwarder-net-id 000042 --forwarder-tenant-id tenant-a \
    --forwarder-cluster-id eu1 --group debug
```

To subscribe to routed uplink traffic as network, tenant, and with or without named cluster:

```bash
$ pbsub --home-network-net-id 000042 --group debug
$ pbsub --home-network-net-id 000042 --home-network-tenant-id tenant-a --group debug
$ pbsub --home-network-net-id 000042 --home-network-cluster-id eu1 --group debug
$ pbsub --home-network-net-id 000042 --home-network-tenant-id tenant-a \
    --home-network-cluster-id eu1 --group debug
```

>**Important**: When using `pbsub`, specify a shared subscription group that is different from the group used in production. Otherwise, traffic gets split to your production subscriptions and your testing subscriptions.

To publish an uplink message in `uplink.json` as Forwarder network, tenant, and with or without named cluster:

```bash
$ pbpub --forwarder-net-id 000042 < uplink.json
$ pbpub --forwarder-net-id 000042 --forwarder-tenant-id tenant-a < uplink.json
$ pbpub --forwarder-net-id 000042 --forwarder-cluster-id eu1 < uplink.json
$ pbpub --forwarder-net-id 000042 --forwarder-tenant-id tenant-a \
    --forwarder-cluster-id eu1 < uplink.json
```

To publish a downlink message in `downlink.json`, as Home Network network, tenant, and with or without named cluster:

```bash
$ pbpub --home-network-net-id 000042 < downlink.json
$ pbpub --home-network-net-id 000042 --home-network-tenant-id tenant-a < downlink.json
$ pbpub --home-network-net-id 000042 --home-network-cluster-id eu1 < downlink.json
$ pbpub --home-network-net-id 000042 --home-network-tenant-id tenant-a \
    --home-network-cluster-id eu1 < downlink.json
```

See [Examples](./examples) for example JSON files.

## Legal

Packet Broker Clients are Apache 2.0 licensed. See [LICENSE](./LICENSE) for more information.
