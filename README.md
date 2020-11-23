# Packet Broker Clients

Packet Broker Clients are command-line utilities for working with [Packet Broker](https://www.packetbroker.org).

---

## Getting Started

### Installation

Make sure you have [Go](https://golang.org/doc/install) installed in your environment.

```bash
$ go get go.packetbroker.org/pb/cmd/pbctl
$ go get go.packetbroker.org/pb/cmd/pbpub
$ go get go.packetbroker.org/pb/cmd/pbsub
```

### Configuration

In order to use the Packet Broker Clients, you need a client certificate signed by Packet Broker CA. Please see [Configuration](./configs) for more information.

Make sure you have `cert.pem`, `key.pem` and [`ca.pem`](./configs/ca.pem) in your working directory.

Instead of passing the Packet Broker address via the `--address` flag on each command, you can set `PB_ADDRESS` in your environment like so:

```bash
export PB_ADDRESS=staging.packetbroker.io
```

If you don't specify a port, `pbadmin`, `pbctl`, `pbpub` and `pbsub` use the default ports:

| Service | Port | Used By |
| --- | ---: | --- |
| IAM | `1900` | `pbadmin` |
| Control Plane | `1900` | `pbctl` |
| Data Plane | `1900` | `pbpub`, `pbsub` |

### Manage Tenants

Packet Broker supports multi-tenancy to assign `DevAddr` blocks to tenants. When routing uplink traffic, Packet Broker Router looks up the Home Network tenant by DevAddr and applies routing policies on the tenant-level. Forwarders can specify a tenant ID as well.

Tenants are optional. When there are no tenants or when a DevAddr does not match any tenant's DevAddr prefixes, the tenant ID remains empty and the routing policies of the NetID is used.

As NetID `000013`, to create or update tenant `tenant-a` with DevAddr prefixes `26AA0000/16` and `26BB0000/16`:

```bash
$ pbctl tenant set --net-id 000013 --tenant-id tenant-a \
    --dev-addr-prefixes 26AA0000/16,26BB0000/16
```

>The prefixes indicate the base DevAddr and the bit length, like [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation). For example, `26AA0000/16` matches all DevAddr from `26AA0000` to `26AAFFFF`.

To list tenants:

```bash
$ pbctl tenant list --net-id 000013
```

To get a tenant:

```bash
$ pbctl tenant get --net-id 000013 --tenant-id tenant-a
```

To delete a tenant:

```bash
$ pbctl tenant delete --net-id 000013 --tenant-id tenant-a
```

### Configure Routing Policies

As a Forwarder, you can configure a default routing policy for all Home Networks, and routing policies per Home Network with `pbctl`. 

As Forwarder NetID `000042`, to see the default routing policy:

```bash
$ pbctl policy get --forwarder-net-id 000042 --defaults
```

To see the routing policy for Home Network NetID `C00123`:

```bash
$ pbctl policy get --forwarder-net-id 000042 --home-network-net-id C00123
```

To see the routing policy for you own network:

```bash
$ pbctl policy get --forwarder-net-id 000042 --home-network-net-id 000042
```

To see the routing policy of Forwrader tenant `tenant-a` for Home Network NetID `C00123` tenant `tenant-b`:

```bash
$ pbctl policy get --forwarder-net-id 000042 --forwarder-tenant-id tenant-a \
    --home-network-net-id C00123 --home-network-tenant-id tenant-b
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

To unset the routing policy, use `--unset`.

### Publish and Subscribe Traffic

To subscribe to routed downlink traffic:

```bash
$ pbsub --forwarder-net-id 000042 --forwarder-id example --group debug
$ pbsub --forwarder-net-id 000042 --forwarder-id example \
    --forwarder-tenant-id test --group debug
```

To subscribe to routed uplink traffic:

```bash
$ pbsub --home-network-net-id 000042 --group debug
$ pbsub --home-network-net-id 000042 --home-network-tenant-id test --group debug
```

You can also subscribe to routed uplink traffic from your own Forwarder NetID and optionally a specific ID:

```bash
$ pbsub --home-network-net-id 000042 --filter-forwarder-net-id 000042 \
    --filter-forwarder-id example --group debug
```

>**Important**: When using `pbsub`, specify a shared subscription group that is different from the group used in production. Otherwise, traffic gets split to your production subscriptions and your testing subscriptions.

To publish an uplink message for testing, you can pipe a JSON file to `pbpub`, specifying a Forwarder:

```bash
$ cat uplink.json | pbpub --forwarder-net-id 000042 --forwarder-id example
$ cat uplink.json | pbpub --forwarder-net-id 000042 --forwarder-id example \
    --forwarder-tenant-id test
```

To publish a downlink message for testing, you can pipe a JSON file to `pbpub`, specifying a Home Network:

```bash
$ cat downlink.json | pbpub --home-network-net-id 000042
$ cat downlink.json | pbpub --home-network-net-id 000042 \
    --home-network-tenant-id test
```

See [Examples](./examples) for example JSON files.

## Legal

Packet Broker Clients are Apache 2.0 licensed. See [LICENSE](./LICENSE) for more information.
